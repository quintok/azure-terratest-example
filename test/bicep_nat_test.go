package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/shell"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

// type expandInstanceViewPolicy struct {
// }

// func (m *expandInstanceViewPolicy) Do(req *policy.Request) (*http.Response, error) {
// 	q := req.Raw().URL.Query()
// 	q.Add("$expand", "instanceView")
// 	req.Raw().URL.RawQuery = q.Encode()
// 	return req.Next()
// }

func getCommand(args string) shell.Command {
	return shell.Command{
		Command: "az",
		Args:    strings.Split(args, " "),
	}
}

func TestNatConfigurationInBicep(t *testing.T) {
	t.Parallel()

	workingDir := "./fixtures/nat-bicep"

	cliLoginIfNotLoggedIn(t)

	defer test_structure.RunTestStage(t, "cleanup", func() {
		rgn := test_structure.LoadString(t, workingDir, "resourceGroupName")
		shell.RunCommand(t, getCommand(fmt.Sprintf("group delete --name %s --yes", rgn)))
	})

	test_structure.RunTestStage(t, "deploy", func() {
		rgn := random.UniqueId()

		test_structure.SaveString(t, workingDir, "resourceGroupName", rgn)
		test_structure.SaveString(t, workingDir, "subscriptionId", os.Getenv("AZURE_SUBSCRIPTION_ID"))
		approvedRegions := []string{
			"eastus",
			"australeast",
		}
		location := azure.GetRandomStableRegion(t, approvedRegions, nil, os.Getenv("AZURE_SUBSCRIPTION_ID"))
		test_structure.SaveString(t, workingDir, "location", location)

		password := "BPN6erd-kjwkf4uen" // this should potentially be generated.
		shell.RunCommand(t, getCommand(fmt.Sprintf("group create --name %s --location %s", rgn, location)))
		shell.RunCommand(t, getCommand(fmt.Sprintf("deployment group create -g %s --template-file fixtures/nat-bicep/main.bicep -p location=%s adminPassword=%s", rgn, location, password)))
	})

	test_structure.RunTestStage(t, "test", func() {
		test_structure.SaveString(t, workingDir, "machineName", "test-machine")
		ValidateNATIsWorking(t, workingDir)
	})
}

func cliLoginIfNotLoggedIn(t *testing.T) {
	out := shell.RunCommandAndGetOutput(t, shell.Command{
		Command: "az",
		Args:    []string{"account", "list"},
	})

	if out == "WARNING: Please run \"az login\" to access your accounts." {
		tenant := os.Getenv("AZURE_TENANT_ID")
		client := os.Getenv("AZURE_CLIENT_ID")
		secret := os.Getenv("AZURE_CLIENT_SECRET")
		subscription := os.Getenv("AZURE_SUBSCRIPTION_ID")

		shell.RunCommand(t, getCommand(fmt.Sprintf("login --service-principal -u %s -p %s --tenant %s", client, secret, tenant)))
		shell.RunCommand(t, getCommand(fmt.Sprintf("account set --subscription %s", subscription)))
	} else {
		logger.Default.Logf(t, "cli already signed in")
	}

}
