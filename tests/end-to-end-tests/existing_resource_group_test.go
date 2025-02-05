package e2e_tests

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

type TestExistingResourceGroupExternalResources struct {
	ResourceGroup armresources.ResourceGroup
}

/*
 * Creates resources which are "external" to the az-backup module.
 */
func setupExternalResourcesForExistingResourceGroupTest(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, resourceGroupLocation string) *TestExistingResourceGroupExternalResources {
	resourceGroup := CreateResourceGroup(t, credential, subscriptionID, resourceGroupName, resourceGroupLocation)

	externalResources := &TestExistingResourceGroupExternalResources{
		ResourceGroup: resourceGroup,
	}

	return externalResources
}

/*
 * TestExistingResourceGroup tests the deployment of a backup vault into a pre-existing resource group.
 */
func TestExistingResourceGroup(t *testing.T) {
	t.Parallel()

	environment := GetEnvironmentConfiguration(t)
	credential := GetAzureCredential(t, environment)

	uniqueId := random.UniqueId()
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", uniqueId)
	resourceGroupLocation := "uksouth"
	backupVaultName := fmt.Sprintf("bvault-nhsbackup-%s", uniqueId)

	externalResources := setupExternalResourcesForExistingResourceGroupTest(t, credential, environment.SubscriptionID, resourceGroupName, resourceGroupLocation)

	// Teardown stage
	// ...

	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, environment.TerraformFolder)

		terraform.Destroy(t, terraformOptions)

		DeleteResourceGroup(t, credential, environment.SubscriptionID, *externalResources.ResourceGroup.Name)
	})

	// Setup stage
	// ...

	test_structure.RunTestStage(t, "setup", func() {
		terraformOptions := &terraform.Options{
			TerraformDir: environment.TerraformFolder,

			Vars: map[string]interface{}{
				"resource_group_name":     resourceGroupName,
				"resource_group_location": resourceGroupLocation,
				"create_resource_group":   false,
				"backup_vault_name":       backupVaultName,
			},

			BackendConfig: map[string]interface{}{
				"resource_group_name":  environment.TerraformStateResourceGroup,
				"storage_account_name": environment.TerraformStateStorageAccount,
				"container_name":       environment.TerraformStateContainer,
				"key":                  backupVaultName + ".tfstate",
			},
		}

		// Save options for later test stages
		test_structure.SaveTerraformOptions(t, environment.TerraformFolder, terraformOptions)

		terraform.InitAndApply(t, terraformOptions)
	})

	// Validate stage
	// ...

	test_structure.RunTestStage(t, "validate", func() {
		// Validate resource group
		resourceGroup := GetResourceGroup(t, environment.SubscriptionID, credential, resourceGroupName)
		assert.NotNil(t, resourceGroup, "Resource group does not exist")
		assert.Equal(t, resourceGroupName, *resourceGroup.Name, "Resource group name does not match")
		assert.Equal(t, resourceGroupLocation, *resourceGroup.Location, "Resource group location does not match")
	})
}
