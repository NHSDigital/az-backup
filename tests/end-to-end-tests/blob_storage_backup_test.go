package e2e_tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

type TestBlobStorageBackupExternalResources struct {
	ResourceGroup     armresources.ResourceGroup
	StorageAccountOne armstorage.Account
	StorageAccountTwo armstorage.Account
}

/*
 * Creates resources which are "external" to the az-backup module, and models
 * what would be backed up in a real scenario.
 */
func setupExternalResourcesForBlobStorageBackupTest(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, vault_name string, vault_location string) *TestBlobStorageBackupExternalResources {
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s-external", vault_name)
	resourceGroup := CreateResourceGroup(t, subscriptionID, credential, resourceGroupName, vault_location)

	storageAccountOneName := fmt.Sprintf("sa%sexternal1", strings.ToLower(vault_name))
	storageAccountOne := CreateStorageAccount(t, credential, subscriptionID, resourceGroupName, storageAccountOneName, vault_location)

	storageAccountTwoName := fmt.Sprintf("sa%sexternal2", strings.ToLower(vault_name))
	storageAccountTwo := CreateStorageAccount(t, credential, subscriptionID, resourceGroupName, storageAccountTwoName, vault_location)

	externalResources := &TestBlobStorageBackupExternalResources{
		ResourceGroup:     resourceGroup,
		StorageAccountOne: storageAccountOne,
		StorageAccountTwo: storageAccountTwo,
	}

	return externalResources
}

/*
 * TestBlobStorageBackup tests the deployment of a backup vault and backup policies for blob storage accounts.
 */
func TestBlobStorageBackup(t *testing.T) {
	t.Parallel()

	environment := GetEnvironmentConfiguration(t)
	credential := GetAzureCredential(t, environment)

	vaultName := random.UniqueId()
	vaultLocation := "uksouth"
	vaultRedundancy := "LocallyRedundant"
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", vaultName)
	backupVaultName := fmt.Sprintf("bvault-%s", vaultName)

	tags := map[string]string{
		"environment":     "production",
		"cost_code":       "code_value",
		"created_by":      "creator_name",
		"created_date":    "01/01/2024",
		"tech_lead":       "tech_lead_name",
		"requested_by":    "requester_name",
		"service_product": "product_name",
		"team":            "team_name",
		"service_level":   "gold",
	}

	externalResources := setupExternalResourcesForBlobStorageBackupTest(t, credential, environment.SubscriptionID, vaultName, vaultLocation)

	// A map of backups which we'll use to apply the TF module, and then validate the
	// policies have been created correctly
	blobStorageBackups := map[string]map[string]interface{}{
		"backup1": {
			"backup_name":        "blob1",
			"retention_period":   "P7D",
			"storage_account_id": *externalResources.StorageAccountOne.ID,
		},
		"backup2": {
			"backup_name":        "blob2",
			"retention_period":   "P30D",
			"storage_account_id": *externalResources.StorageAccountTwo.ID,
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
				"vault_name":           vaultName,
				"vault_location":       vaultLocation,
				"vault_redundancy":     vaultRedundancy,
				"tags":                 tags,
				"blob_storage_backups": blobStorageBackups,
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
		backupVault := GetBackupVault(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName)
		backupPolicies := GetBackupPolicies(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName)
		backupInstances := GetBackupInstances(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName)

		assert.Equal(t, len(blobStorageBackups), len(backupPolicies), "Expected to find %2 backup policies in vault", len(blobStorageBackups))
		assert.Equal(t, len(blobStorageBackups), len(backupInstances), "Expected to find %2 backup instances in vault", len(blobStorageBackups))

		for _, backup := range blobStorageBackups {
			backupName := backup["backup_name"].(string)
			retentionPeriod := backup["retention_period"].(string)
			storageAccountId := backup["storage_account_id"].(string)

			// Validate backup policy
			backupPolicyName := fmt.Sprintf("bkpol-%s-blobstorage-%s", vaultName, backupName)
			backupPolicy := GetBackupPolicyForName(backupPolicies, backupPolicyName)
			assert.NotNil(t, backupPolicy, "Expected to find a backup policy called %s", backupPolicyName)

			// Validate retention period
			backupPolicyProperties := backupPolicy.Properties.(*armdataprotection.BackupPolicy)
			retentionRule := GetBackupPolicyRuleForName(backupPolicyProperties.PolicyRules, "Default").(*armdataprotection.AzureRetentionRule)
			deleteOption := retentionRule.Lifecycles[0].DeleteAfter.(*armdataprotection.AbsoluteDeleteOption)
			assert.Equal(t, retentionPeriod, *deleteOption.Duration, "Expected the backup policy retention period to be %s", retentionPeriod)

			// Validate backup instance
			backupInstanceName := fmt.Sprintf("bkinst-%s-blobstorage-%s", vaultName, backupName)
			backupInstance := GetBackupInstanceForName(backupInstances, backupInstanceName)
			assert.NotNil(t, backupInstance, "Expected to find a backup policy called %s", backupInstanceName)
			assert.Equal(t, storageAccountId, *backupInstance.Properties.DataSourceInfo.ResourceID, "Expected the backup instance source resource ID to be %s", storageAccountId)
			assert.Equal(t, *backupPolicy.ID, *backupInstance.Properties.PolicyInfo.PolicyID, "Expected the backup instance policy ID to be %s", backupPolicy.ID)

			// Validate role assignment
			backupContributorRoleDefinition := GetRoleDefinition(t, credential, "Storage Account Backup Contributor")
			backupContributorRoleAssignment := GetRoleAssignment(t, credential, environment.SubscriptionID, *backupVault.Identity.PrincipalID, backupContributorRoleDefinition, storageAccountId)
			assert.NotNil(t, backupContributorRoleAssignment, "Expected to find role assignment %s for principal %s on scope %s", backupContributorRoleDefinition.Name, *backupVault.Identity.PrincipalID, storageAccountId)
		}
	})
}
