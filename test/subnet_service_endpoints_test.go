package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-12-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformServiceEndpoints(t *testing.T) {
	t.Parallel()

	terraformOptions := configureTerraformOptions(t)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")
	vnetName := terraform.Output(t, terraformOptions, "vnet_name")
	subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
	assert.NotEmpty(t, azure.GetVirtualNetworkSubnets(t, vnetName, resourceGroupName, subscriptionId))

	// 1. Build a network
	// 2. Provision a VM
	// 3. Run

	client := compute.NewVirtualMachineRunCommandsClient(subscriptionId)
	future, err := client.CreateOrUpdate(context.Background(), "examplerg", "mycomputer2", "ifconfig", compute.VirtualMachineRunCommand{
		VirtualMachineRunCommandProperties: &compute.VirtualMachineRunCommandProperties{
			AsyncExecution: to.BoolPtr(false),
			Source: &compute.VirtualMachineRunCommandScriptSource{
				Script: to.StringPtr("ifconfig"),
			},
		},
	})
	if err != nil {
		require.NoError(t, err)
	}

	err = future.WaitForCompletionRef(context.Background(), client.Client)
	if err != nil {
		require.NoError(t, err)
	}

	res, err := future.Result(client)
	if err != nil {
		require.NoError(t, err)
	}
	fmt.Println(json.Marshal(res))

}

func configureTerraformOptions(t *testing.T) *terraform.Options {

	uniquePostfix := random.UniqueId()
	terraformOptions := &terraform.Options{
		TerraformDir: "../",
		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"postfix": uniquePostfix,
		},
	}

	return terraformOptions
}
