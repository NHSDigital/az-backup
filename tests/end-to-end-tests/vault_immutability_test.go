package e2e_tests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

type TestVaultImmutabilityExternalResources struct {
	ResourceGroup           armresources.ResourceGroup
	StorageAccount          armstorage.Account
	StorageAccountContainer armstorage.BlobContainer
}

/*
 * Creates resources which are "external" to the az-backup module, and models
 * what would be backed up in a real scenario.
 */
func setupExternalResourcesForVaultImmutabilityTest(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, resourceGroupLocation string, uniqueId string) *TestVaultImmutabilityExternalResources {
	externalResourceGroupName := fmt.Sprintf("%s-external", resourceGroupName)
	resourceGroup := CreateResourceGroup(t, credential, subscriptionID, externalResourceGroupName, resourceGroupLocation)

	storageAccountName := fmt.Sprintf("sa%sexternal", strings.ToLower(uniqueId))
	storageAccount := CreateStorageAccount(t, credential, subscriptionID, externalResourceGroupName, storageAccountName, resourceGroupLocation)
	storageAccountContainer := CreateStorageAccountContainer(t, credential, subscriptionID, externalResourceGroupName, storageAccountName, "test-container")

	externalResources := &TestVaultImmutabilityExternalResources{
		ResourceGroup:           resourceGroup,
		StorageAccount:          storageAccount,
		StorageAccountContainer: storageAccountContainer,
	}

	return externalResources
}

/*
 * TestVaultImmutability tests the immutability of the backup vault.
 */
func TestVaultImmutability(t *testing.T) {
	t.Parallel()

	environment := GetEnvironmentConfiguration(t)
	credential := GetAzureCredential(t, environment)

	uniqueId := "b1Watt" //random.UniqueId()
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", uniqueId)
	resourceGroupLocation := "uksouth"
	backupVaultName := fmt.Sprintf("bvault-nhsbackup-%s", uniqueId)
	backupVaultImmutability := "Unlocked"

	externalResources := setupExternalResourcesForVaultImmutabilityTest(t, credential, environment.SubscriptionID, resourceGroupName, resourceGroupLocation, uniqueId)

	// A map of backups which we'll use to apply the TF module, and then validate the
	// policies have been created correctly
	blobStorageBackups := map[string]map[string]interface{}{
		"backup1": {
			"backup_name":                "blob1",
			"retention_period":           "P7D",
			"backup_intervals":           []string{"R/2024-01-01T00:00:00+00:00/P1D"},
			"storage_account_id":         *externalResources.StorageAccount.ID,
			"storage_account_containers": []string{*externalResources.StorageAccountContainer.Name},
		},
	}

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
				"resource_group_name":       resourceGroupName,
				"resource_group_location":   resourceGroupLocation,
				"backup_vault_name":         backupVaultName,
				"backup_vault_immutability": backupVaultImmutability,
				"blob_storage_backups":      blobStorageBackups,
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
		testFile := CreateTestFile(t)
		defer os.Remove(testFile.Name())

		UploadFileToStorageAccount(t, credential, environment.SubscriptionID, *externalResources.ResourceGroup.Name,
			*externalResources.StorageAccount.Name, *externalResources.StorageAccountContainer.Name, testFile.Name())

		backupInstanceName := fmt.Sprintf("bkinst-blob-%s", blobStorageBackups["backup1"]["backup_name"].(string))
		BeginAdHocBackup(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName, backupInstanceName)

		errOne := DeleteBackupInstance(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName, backupInstanceName)
		assert.Error(t, errOne, "Expected an error when deleting a backup instance from an immutable vault: %v", errOne)

		disabledState := armdataprotection.ImmutabilityStateDisabled
		UpdateBackupVaultImmutability(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName, armdataprotection.ImmutabilitySettings{
			State: &disabledState,
		})

		errTwo := DeleteBackupInstance(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName, backupInstanceName)
		assert.NoError(t, errTwo, "Expected no error when deleting a backup instance from an unlocked vault: %v", errTwo)
	})
}
