package test

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
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
