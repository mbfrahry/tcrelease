package cli

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/spf13/cobra"
)

type TCFlags struct {
	ServerUrl   string
	BuildTypeId string
	User        string
	Pass        string
}

type FlagData struct {
	TC TCFlags
}

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
		Use:   "tcrelease provider [test regex]",
		Short: "tcrelease is a small utility to trigger provider releases on teamcity",
		Args:    cobra.ExactArgs(3),
		PreRunE: ValidateParams([]string{"server", "buildtypeid", "user"}),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			targetVersion := args[1]
			nextVersion := args[2]

			// at this point command validation has been done so any more errors dont' require help to be printed
			cmd.SilenceErrors = true

			return TcCmd(viper.GetString("server"), viper.GetString("buildtypeid"), provider, targetVersion, nextVersion, viper.GetString("user"), viper.GetString("pass"))
		},
	}

	pflags := root.PersistentFlags()
	pflags.StringVarP(&flags.TC.ServerUrl, "server", "s", "", "the TeamCity server's uri")
	pflags.StringVarP(&flags.TC.BuildTypeId, "buildtypeid", "b", "", "the TeamCity BuildTypeId to trigger")
	pflags.StringVarP(&flags.TC.User, "user", "u", "", "the TeamCity user to use")
	pflags.StringVarP(&flags.TC.Pass, "pass", "p", "", "the TeamCity password to use (consider exporting pass to TCTEST_PASS instead)")

	viper.BindPFlag("server", pflags.Lookup("server"))
	viper.BindPFlag("buildtypeid", pflags.Lookup("buildtypeid"))
	viper.BindPFlag("user", pflags.Lookup("user"))
	viper.BindPFlag("pass", pflags.Lookup("pass"))

	viper.BindPFlag("provider", pflags.Lookup("provider"))

	viper.BindEnv("server", "TCTEST_SERVER")
	viper.BindEnv("buildtypeid", "TCTEST_BUILDTYPEID")
	viper.BindEnv("user", "TCTEST_USER")
	viper.BindEnv("pass", "TCTEST_PASS")

	viper.BindEnv("provider", "TCTEST_PROVIDER")

	return root
}
