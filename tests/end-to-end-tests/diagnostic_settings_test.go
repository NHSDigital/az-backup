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

type TestDiagnosticSettingsExternalResources struct {
	ResourceGroup         armresources.ResourceGroup
	LogAnalyticsWorkspace armoperationalinsights.Workspace
}

/*
 * Creates resources which are "external" to the az-backup module, and models
 * what would be backed up in a real scenario.
 */
func setupExternalResourcesForDiagnosticSettingsTest(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, resourceGroupLocation string, uniqueId string) *TestDiagnosticSettingsExternalResources {
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
 * TestDiagnosticSettings tests the configuration of the backup vaults diagnostics settings and ensures they
 * integrate with an external log analytics workspace.
 */
func TestDiagnosticSettings(t *testing.T) {
	t.Parallel()

	environment := GetEnvironmentConfiguration(t)
	credential := GetAzureCredential(t, environment)

	uniqueId := random.UniqueId()
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", uniqueId)
	resourceGroupLocation := "uksouth"
	backupVaultName := fmt.Sprintf("bvault-nhsbackup-%s", uniqueId)
	backupVaultRedundancy := "LocallyRedundant"

	externalResources := setupExternalResourcesForDiagnosticSettingsTest(t, credential, environment.SubscriptionID, resourceGroupName, resourceGroupLocation, uniqueId)

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

		// Save options for later test stages
		test_structure.SaveTerraformOptions(t, environment.TerraformFolder, terraformOptions)

		terraform.InitAndApply(t, terraformOptions)
	})

	// Validate stage
	// ...

	test_structure.RunTestStage(t, "validate", func() {
		// An array of log categories that we expect to be enabled for the diagnostic settings
		expectedLogCategories := []string{
			"AddonAzureBackupJobs",
			"AddonAzureBackupPolicy",
			"AddonAzureBackupProtectedInstance",
			"CoreAzureBackup",
		}

		// An array of metrics that we expect to be enabled for the diagnostic settings
		expectedMetricCategories := []string{
			"Health",
		}

		backupVault := GetBackupVault(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName)
		diagnosticSettings := GetDiagnosticSettings(t, credential, *backupVault.ID, *backupVault.Name)

		assert.Equal(t, len(diagnosticSettings.Properties.Logs), len(expectedLogCategories), "Expected to find %2 log categories in diagnostic settings", len(expectedLogCategories))
		assert.Equal(t, len(diagnosticSettings.Properties.Metrics), len(expectedMetricCategories), "Expected to find %2 metric categories in diagnostic settings", len(expectedMetricCategories))

		for _, expectedCategory := range expectedLogCategories {
			found := false
			for _, log := range diagnosticSettings.Properties.Logs {
				if *log.Category == expectedCategory {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected log category %s not found in diagnostic settings", expectedCategory)
		}

		for _, expectedCategory := range expectedMetricCategories {
			found := false
			for _, metric := range diagnosticSettings.Properties.Metrics {
				if *metric.Category == expectedCategory {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected metric category %s not found in diagnostic settings", expectedCategory)
		}
	})
}
