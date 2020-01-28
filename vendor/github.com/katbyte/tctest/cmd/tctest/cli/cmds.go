package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/katbyte/tctest/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type TCFlags struct {
	ServerUrl   string
	BuildTypeId string
	User        string
	Pass        string
}

type PRFlags struct {
	Repo      string
	FileRegEx string
	TestSplit string
}

type FlagData struct {
	TC TCFlags
	PR PRFlags
	ServicePackagesMode bool
}

// colours
// PR - cyan
// urls dim blue
// tests - pink/purple

// OUTPUT
//discovering tests (github url to PR)
// test1 colour?)
// test2
//triggering build_id(white) @ BRANCH(white) with PATTERN(white)...
//  started build dim green) #123(bright green) (url to build) (dim)
//if wait, live update buildlog every x seconds

func ValidateParams(params []string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, p := range params {
			if viper.GetString(p) == "" {
				return fmt.Errorf(p + " paramter can't be empty")
			}
		}

		return nil
	}
}

func Make() *cobra.Command {

	flags := FlagData{}

	root := &cobra.Command{
		Use:   "tctest branch [test regex]",
		Short: "tctest is a small utility to trigger acceptance tests on teamcity",
		Long: `A small utility to trigger acceptance tests on teamcity. 
It can also pull the tests to run for a PR on github
Complete documentation is available at https://github.com/katbyte/tctest`,
		Args:    cobra.ExactArgs(2),
		PreRunE: ValidateParams([]string{"server", "buildtypeid", "user"}),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch := args[0]
			testRegEx := args[1]

			if !strings.HasPrefix(branch, "refs/") {
				branch = "refs/heads/" + branch
			}

			// at this point command validation has been done so any more errors dont' require help to be printed
			cmd.SilenceErrors = true

			return TcCmd(viper.GetString("server"), viper.GetString("buildtypeid"), branch, testRegEx, viper.GetString("user"), viper.GetString("pass"))
		},
	}

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of tctest",
		Long:  `Print the version number of tctest`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("tctest v" + version.Version + "-" + version.GitCommit)
		},
	})

	pr := &cobra.Command{
		Use:     "pr # [test_regex]",
		Short:   "triggers acceptance tests matching regex for a PR",
		Long:    `TODO`,
		Args:    cobra.RangeArgs(1, 2),
		PreRunE: ValidateParams([]string{"server", "buildtypeid", "user", "repo", "fileregex", "splittests"}),
		RunE: func(cmd *cobra.Command, args []string) error {
			pr := args[0]
			testRegEx := ""
			repo := viper.GetString("repo")

			if len(args) == 2 {
				testRegEx = args[1]
			}
			if _, err := strconv.Atoi(pr); err != nil {
				return fmt.Errorf("pr should be a number: %v", err)
			}

			// at this point command validation has been done so any more errors dont' require help to be printed
			cmd.SilenceErrors = true

			state, err := PrState(repo, pr)
			if err != nil {
				return fmt.Errorf("unable to get pr state: %v", err)
			}
			if state == "closed" {
				return fmt.Errorf("cannot start build for a closed pr")
			}

			if testRegEx == "" {
				tests, err := PrCmd(viper.GetString("repo"), pr, viper.GetString("fileregex"), viper.GetString("splittests"), viper.GetBool("servicepackages"))
				if err != nil {
					return fmt.Errorf("pr cmd failed: %v", err)
				}

				testRegEx = "(" + strings.Join(*tests, "|") + ")"

			}

			branch := fmt.Sprintf("refs/pull/%s/merge", pr)
			return TcCmd(viper.GetString("server"), viper.GetString("buildtypeid"), branch, testRegEx, viper.GetString("user"), viper.GetString("pass"))
		},
	}
	root.AddCommand(pr)

	list := &cobra.Command{
		Use:     "list #",
		Short:   "attempts to discover what acceptance tests to run for a PR",
		Long:    `TODO`,
		Args:    cobra.RangeArgs(1, 1),
		PreRunE: ValidateParams([]string{"repo", "fileregex", "splittests"}),
		RunE: func(cmd *cobra.Command, args []string) error {
			pr := args[0]

			// at this point command validation has been done so any more errors dont' require help to be printed
			cmd.SilenceErrors = true

			if _, err := PrCmd(viper.GetString("repo"), pr, viper.GetString("fileregex"), viper.GetString("splittests"), viper.GetBool("servicepackages")); err != nil {
				return fmt.Errorf("pr cmd failed: %v", err)
			}

			return nil
		},
	}
	pr.AddCommand(list)

	status := &cobra.Command{
		Use:     "status #",
		Short:   "shows the test status for a specifed TC build ID",
		Long:    "shows the test status for a specifed TC build ID",
		Args:    cobra.RangeArgs(1, 1),
		PreRunE: ValidateParams([]string{"server", "user"}),
		RunE: func(cmd *cobra.Command, args []string) error {
			buildId := args[0]
			return TcTestStatus(viper.GetString("server"), buildId, viper.GetString("user"), viper.GetString("pass"))
		},
	}
	root.AddCommand(status)

	pflags := root.PersistentFlags()
	pflags.StringVarP(&flags.TC.ServerUrl, "server", "s", "", "the TeamCity server's ur;")
	pflags.StringVarP(&flags.TC.BuildTypeId, "buildtypeid", "b", "", "the TeamCity BuildTypeId to trigger")
	pflags.StringVarP(&flags.TC.User, "user", "u", "", "the TeamCity user to use")
	pflags.StringVarP(&flags.TC.Pass, "pass", "p", "", "the TeamCity password to use (consider exporting pass to TCTEST_PASS instead)")

	pflags.StringVarP(&flags.PR.Repo, "repo", "r", "", "repository the pr resides in, such as terraform-providers/terraform-provider-azurerm")
	pflags.StringVarP(&flags.PR.FileRegEx, "fileregex", "", "(^[a-z]*/resource_|^[a-z]*/data_source_)", "the regex to filter files by`")
	pflags.StringVar(&flags.PR.TestSplit, "splittests", "_", "split tests here and use the value on the left")

	pflags.BoolVar(&flags.ServicePackagesMode, "servicepackages", false, "enable service packages mode for AzureRM")


	viper.BindPFlag("server", pflags.Lookup("server"))
	viper.BindPFlag("buildtypeid", pflags.Lookup("buildtypeid"))
	viper.BindPFlag("user", pflags.Lookup("user"))
	viper.BindPFlag("pass", pflags.Lookup("pass"))

	viper.BindPFlag("repo", pflags.Lookup("repo"))
	viper.BindPFlag("fileregex", pflags.Lookup("fileregex"))
	viper.BindPFlag("splittests", pflags.Lookup("splittests"))

	viper.BindPFlag("servicepackages", pflags.Lookup("servicepackages"))

	viper.BindEnv("server", "TCTEST_SERVER")
	viper.BindEnv("buildtypeid", "TCTEST_BUILDTYPEID")
	viper.BindEnv("user", "TCTEST_USER")
	viper.BindEnv("pass", "TCTEST_PASS")

	viper.BindEnv("repo", "TCTEST_REPO")
	viper.BindEnv("fileregex", "TCTEST_FILEREGEX")
	viper.BindEnv("splittests", "TCTEST_SPLITTESTS")

	viper.BindEnv("servicepackages", "TCTEST_SERVICEPACKAGESMODE")

	//todo config file
	/*viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath("/etc/appname/")   // path to look for the config file in
	viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}*/

	return root
}
