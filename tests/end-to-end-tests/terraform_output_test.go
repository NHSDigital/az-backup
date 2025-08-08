package e2e_tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

type TestTerraformOutputsExternalResources struct {
	ResourceGroup         armresources.ResourceGroup
	LogAnalyticsWorkspace armoperationalinsights.Workspace
}

/*
 * Creates resources which are "external" to the az-backup module, and models
 * what would be backed up in a real scenario.
 */
func setupExternalResourcesForTerraformOutputTest(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, resourceGroupLocation string, uniqueId string) *TestDiagnosticSettingsExternalResources {
	externalResourceGroupName := fmt.Sprintf("%s-external", resourceGroupName)
	resourceGroup := CreateResourceGroup(t, credential, subscriptionID, externalResourceGroupName, resourceGroupLocation)

	logAnalyticsWorkspaceName := fmt.Sprintf("law-%s-external", strings.ToLower(uniqueId))
	logAnalyticsWorkspace := CreateLogAnalyticsWorkspace(t, credential, subscriptionID, externalResourceGroupName, logAnalyticsWorkspaceName, resourceGroupLocation)

	externalResources := &TestDiagnosticSettingsExternalResources{
		ResourceGroup:         resourceGroup,
		LogAnalyticsWorkspace: logAnalyticsWorkspace,
	}

	return externalResources
}

/*
 * TestTerraformOutput tests the output variables of the Terraform deployment.
 */
func TestTerraformOutput(t *testing.T) {
	t.Parallel()

	environment := GetEnvironmentConfiguration(t)
	credential := GetAzureCredential(t, environment)

	uniqueId := random.UniqueId()
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", uniqueId)
	resourceGroupLocation := "uksouth"
	backupVaultName := fmt.Sprintf("bvault-nhsbackup-%s", uniqueId)
	backupVaultRedundancy := "LocallyRedundant"

	externalResources := setupExternalResourcesForTerraformOutputTest(t, credential, environment.SubscriptionID, resourceGroupName, resourceGroupLocation, uniqueId)

	// Teardown stage
	// ...

	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, environment.TerraformFolder)

		terraform.Destroy(t, terraformOptions)
	})

	// Setup stage
	// ...

	test_structure.RunTestStage(t, "setup", func() {
		terraformOptions := &terraform.Options{
			TerraformDir: environment.TerraformFolder,

			Vars: map[string]interface{}{
				"resource_group_name":        resourceGroupName,
				"resource_group_location":    resourceGroupLocation,
				"backup_vault_name":          backupVaultName,
				"backup_vault_redundancy":    backupVaultRedundancy,
				"log_analytics_workspace_id": *externalResources.LogAnalyticsWorkspace.ID,
			},

			BackendConfig: map[string]interface{}{
				"resource_group_name":  environment.TerraformStateResourceGroup,
				"storage_account_name": environment.TerraformStateStorageAccount,
				"container_name":       environment.TerraformStateContainer,
				"key":                  backupVaultName + ".tfstate",
			},
		}

		test_structure.SaveTerraformOptions(t, environment.TerraformFolder, terraformOptions)

		terraform.InitAndApply(t, terraformOptions)
	})

	// Validate stage
	// ...

	test_structure.RunTestStage(t, "validate", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, environment.TerraformFolder)

		actualVaultName := terraform.OutputMap(t, terraformOptions, "backup_vault")["name"]
		assert.Equal(t, backupVaultName, actualVaultName)

		actualVaultLocation := terraform.OutputMap(t, terraformOptions, "backup_vault")["location"]
		assert.Equal(t, resourceGroupLocation, actualVaultLocation)

		actualVaultRedundancy := terraform.OutputMap(t, terraformOptions, "backup_vault")["redundancy"]
		assert.Equal(t, backupVaultRedundancy, actualVaultRedundancy)
	})
}
