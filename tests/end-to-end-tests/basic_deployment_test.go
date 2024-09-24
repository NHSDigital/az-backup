package e2e_tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
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

	vaultName := random.UniqueId()
	vaultLocation := "uksouth"
	vaultRedundancy := "LocallyRedundant"

	// Teardown stage - deferred so it runs after the other test stages
	// regardless of whether they succeed or fail.
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

		// Save options for later test stages
		test_structure.SaveTerraformOptions(t, environment.TerraformFolder, terraformOptions)

		terraform.InitAndApply(t, terraformOptions)
	})

	// Validate stage
	// ...

	test_structure.RunTestStage(t, "validate", func() {
		resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", vaultName)
		fullVaultName := fmt.Sprintf("bvault-%s", vaultName)

		// Create a credential to authenticate with Azure Resource Manager
		cred, err := azidentity.NewClientSecretCredential(environment.TenantID, environment.ClientID, environment.ClientSecret, nil)
		assert.NoError(t, err, "Failed to obtain a credential: %v", err)

		ValidateResourceGroup(t, environment.SubscriptionID, cred, resourceGroupName, vaultLocation)
		ValidateBackupVault(t, environment.SubscriptionID, cred, resourceGroupName, fullVaultName, vaultLocation)
	})
}

/*
 * Validates the resource group has been deployed correctly
 */
func ValidateResourceGroup(t *testing.T, subscriptionID string,
	cred *azidentity.ClientSecretCredential, resourceGroupName string, vaultLocation string) {
	// Create a new resource groups client
	client, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	assert.NoError(t, err, "Failed to create resource group client: %v", err)

	// Get the resource group
	resp, err := client.Get(context.Background(), resourceGroupName, nil)
	assert.NoError(t, err, "Failed to get resource group: %v", err)

	// Validate the resource group
	assert.NotNil(t, resp.ResourceGroup, "Resource group does not exist")
	assert.Equal(t, resourceGroupName, *resp.ResourceGroup.Name, "Resource group name does not match")
	assert.Equal(t, vaultLocation, *resp.ResourceGroup.Location, "Resource group location does not match")
}

/*
 * Validates the backup vault has been deployed correctly
 */
func ValidateBackupVault(t *testing.T, subscriptionID string, cred *azidentity.ClientSecretCredential, resourceGroupName string, vaultName string, vaultLocation string) {
	// Create a new Data Protection Backup Vaults client
	client, err := armdataprotection.NewBackupVaultsClient(subscriptionID, cred, nil)
	assert.NoError(t, err, "Failed to create data protection client: %v", err)

	// Get the backup vault
	resp, err := client.Get(context.Background(), resourceGroupName, vaultName, nil)
	assert.NoError(t, err, "Failed to get backup vault: %v", err)

	// Validate the backup vault
	assert.NotNil(t, resp.BackupVaultResource, "Backup vault does not exist")
	assert.Equal(t, vaultName, *resp.BackupVaultResource.Name, "Backup vault name does not match")
	assert.Equal(t, vaultLocation, *resp.BackupVaultResource.Location, "Backup vault location does not match")
	assert.NotNil(t, resp.BackupVaultResource.Identity.PrincipalID, "Backup vault identity does not exist")
	assert.Equal(t, "SystemAssigned", *resp.BackupVaultResource.Identity.Type, "Backup vault identity type does not match")
	assert.Equal(t, armdataprotection.StorageSettingTypesLocallyRedundant, *resp.BackupVaultResource.Properties.StorageSettings[0].Type, "Backup vault redundancy does not match")
	assert.Equal(t, armdataprotection.StorageSettingStoreTypesVaultStore, *resp.BackupVaultResource.Properties.StorageSettings[0].DatastoreType, "Backup vault datastore type does not match")
}
