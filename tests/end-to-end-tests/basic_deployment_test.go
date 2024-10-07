package e2e_tests

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

/*
 * TestBasicDeployment tests the basic deployment of the infrastructure using Terraform.
 */
func TestBasicDeployment(t *testing.T) {
	t.Parallel()

	environment := GetEnvironmentConfiguration(t)
	credential := GetAzureCredential(t, environment)

	vaultName := random.UniqueId()
	vaultLocation := "uksouth"
	vaultRedundancy := "LocallyRedundant"
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", vaultName)
	backupVaultName := fmt.Sprintf("bvault-%s", vaultName)

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
				"vault_name":       vaultName,
				"vault_location":   vaultLocation,
				"vault_redundancy": vaultRedundancy,
			},

			BackendConfig: map[string]interface{}{
				"resource_group_name":  environment.TerraformStateResourceGroup,
				"storage_account_name": environment.TerraformStateStorageAccount,
				"container_name":       environment.TerraformStateContainer,
				"key":                  vaultName + ".tfstate",
			},
		}

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
		assert.Equal(t, vaultLocation, *resourceGroup.Location, "Resource group location does not match")

		// Validate backup vault
		backupVault := GetBackupVault(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName)
		assert.NotNil(t, backupVault, "Backup vault does not exist")
		assert.Equal(t, backupVaultName, *backupVault.Name, "Backup vault name does not match")
		assert.Equal(t, vaultLocation, *backupVault.Location, "Backup vault location does not match")
		assert.NotNil(t, backupVault.Identity.PrincipalID, "Backup vault identity does not exist")
		assert.Equal(t, "SystemAssigned", *backupVault.Identity.Type, "Backup vault identity type does not match")
		assert.Equal(t, armdataprotection.StorageSettingTypesLocallyRedundant, *backupVault.Properties.StorageSettings[0].Type, "Backup vault redundancy does not match")
		assert.Equal(t, armdataprotection.StorageSettingStoreTypesVaultStore, *backupVault.Properties.StorageSettings[0].DatastoreType, "Backup vault datastore type does not match")
	})
}
