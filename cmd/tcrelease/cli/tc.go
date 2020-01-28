package cli

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"syscall"

	"github.com/katbyte/tctest/common"
	"golang.org/x/crypto/ssh/terminal"
)

func TcCmd(server, buildTypeId, provider, targetVersion, nextVersion, user, pass string) error {
	fmt.Printf("triggering %s for %s...</>\n", provider, targetVersion)
	fmt.Printf("%s@%s#%s\n", user, server, buildTypeId)

	// prompt for password if not passed in somehow
	if pass == "" {
		fmt.Print("  password:")
		passBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("unable to read in password : %v", err)
		}
		pass = string(passBytes)
		fmt.Println("")
	}

	build, buildUrl, err := TcBuild(server, buildTypeId, provider, targetVersion, nextVersion, user, pass)
	if err != nil {
		return fmt.Errorf("unable to trigger build: %v", err)
	}

	fmt.Printf("  build %s queued! (%s)\n", build, buildUrl)

	return nil
}

func TcBuild(server, buildTypeId, provider, targetVersion, nextVersion, user, pass string) (string, string, error) {

	url := fmt.Sprintf("https://%s/app/rest/buildQueue", server)
	body := fmt.Sprintf(`
<build>
	<buildType id="%s"/>
	<properties>
		<property name="env.DEPLOYMENT_TARGET_VERSION" value="%s"/>
		<property name="env.DEPLOYMENT_NEXT_VERSION" value="%s"/>
		<property name="env.PROVIDER_NAME" value="%s"/>
	</properties>
</build>
`, buildTypeId, targetVersion, nextVersion, provider)

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("building build request failed: %v", err)
	}

	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/xml")

	resp, err := common.Http.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("build request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP status NOT OK: %d", resp.StatusCode)
	}

	data := struct {
		BuildId string `xml:"id,attr"`
	}{}
	if err := xml.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", fmt.Errorf("unable to decode XML: %d", resp.StatusCode)
	}

	return data.BuildId, fmt.Sprintf("https://%s/viewQueued.html?itemId=%s", server, data.BuildId), nil
}
