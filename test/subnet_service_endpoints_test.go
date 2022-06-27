package test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type expandInstanceViewPolicy struct {
}

func (m *expandInstanceViewPolicy) Do(req *policy.Request) (*http.Response, error) {
	q := req.Raw().URL.Query()
	q.Add("$expand", "instanceView")
	req.Raw().URL.RawQuery = q.Encode()
	return req.Next()
}

func TestNATConfiguration(t *testing.T) {
	t.Parallel()

	workingDir := "./fixtures/nat"

	// when test complete, destroy fixture.
	defer test_structure.RunTestStage(t, "cleanup_network", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)
		terraform.Destroy(t, terraformOptions)
	})

	// Build fixture
	test_structure.RunTestStage(t, "network", func() {
		test_structure.SaveString(t, workingDir, "resourceGroupName", random.UniqueId())
		test_structure.SaveString(t, workingDir, "subscriptionId", os.Getenv("AZURE_SUBSCRIPTION_ID"))
		approvedRegions := []string{
			"eastus",
			"australeast",
		}
		test_structure.SaveString(t, workingDir, "location", azure.GetRandomStableRegion(t, approvedRegions, nil, os.Getenv("AZURE_SUBSCRIPTION_ID")))

		terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
			TerraformDir: workingDir,
			Vars: map[string]interface{}{
				"location":            test_structure.LoadString(t, workingDir, "location"),
				"resource_group_name": test_structure.LoadString(t, workingDir, "resourceGroupName"),
			},
		})

		test_structure.SaveTerraformOptions(t, workingDir, terraformOptions)
		terraform.InitAndApply(t, terraformOptions)

		machineName := terraform.Output(t, terraformOptions, "machine-name")
		test_structure.SaveString(t, workingDir, "machineName", machineName)
	})

	test_structure.RunTestStage(t, "validate", func() {
		ValidateNATIsWorking(t, workingDir)
	})
}

func ValidateNATIsWorking(t *testing.T, workingDir string) {
	ctx := context.Background()

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	require.NoError(t, err)

	options := &arm.ClientOptions{}
	options.PerCallPolicies = []policy.Policy{&expandInstanceViewPolicy{}}

	subscriptionId := test_structure.LoadString(t, workingDir, "subscriptionId")
	resourceGroupname := test_structure.LoadString(t, workingDir, "resourceGroupName")
	machineName := test_structure.LoadString(t, workingDir, "machineName")
	location := test_structure.LoadString(t, workingDir, "location")
	client, err := armcompute.NewVirtualMachineRunCommandsClient(subscriptionId, cred, options)
	require.NoError(t, err)
	pollerRespoonse, err := client.BeginCreateOrUpdate(ctx, resourceGroupname, machineName, "confirm-nat", armcompute.VirtualMachineRunCommand{
		Location: to.Ptr(location),
		Properties: &armcompute.VirtualMachineRunCommandProperties{
			Source: &armcompute.VirtualMachineRunCommandScriptSource{
				Script: to.Ptr("curl -s -o /dev/null -w \"%{http_code}\" google.com"),
			},
		},
	}, nil)
	require.NoError(t, err)
	res, err := pollerRespoonse.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 2 * time.Second,
	})
	require.NoError(t, err)
	assert.Equal(t, "301", *res.Properties.InstanceView.Output)
}
