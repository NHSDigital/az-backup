package e2e_tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

type TestBlobStorageBackupExternalResources struct {
	ResourceGroup               armresources.ResourceGroup
	LogAnalyticsWorkspace       armoperationalinsights.Workspace
	StorageAccountOne           armstorage.Account
	StorageAccountOneContainerA armstorage.BlobContainer
	StorageAccountOneContainerB armstorage.BlobContainer
	StorageAccountTwo           armstorage.Account
	StorageAccountTwoContainerA armstorage.BlobContainer
}

/*
 * Creates resources which are "external" to the az-backup module, and models what would be backed
 * up in a real scenario.
 *
 * The setup is two storage accounts, one of them containing two blob containers. Backups scenarios will be
 * created for:
 * - Backup 1: Storage account one / container A
 * - Backup 2: Storage account one / container B
 * - Backup 3: Storage account one / container A (duplicate of the first scenario, but with a different policy)
 * - Backup 4: Storage account two / container A
 */
func setupExternalResourcesForBlobStorageBackupTest(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, resourceGroupLocation string, uniqueId string) *TestBlobStorageBackupExternalResources {
	externalResourceGroupName := fmt.Sprintf("%s-external", resourceGroupName)
	resourceGroup := CreateResourceGroup(t, credential, subscriptionID, externalResourceGroupName, resourceGroupLocation)

	logAnalyticsWorkspaceName := fmt.Sprintf("law-%s-external", strings.ToLower(uniqueId))
	logAnalyticsWorkspace := CreateLogAnalyticsWorkspace(t, credential, subscriptionID, externalResourceGroupName, logAnalyticsWorkspaceName, resourceGroupLocation)

	storageAccountOneName := fmt.Sprintf("sa%sexternal1", strings.ToLower(uniqueId))
	storageAccountOne := CreateStorageAccount(t, credential, subscriptionID, externalResourceGroupName, storageAccountOneName, resourceGroupLocation)
	storageAccountOneContainerA := CreateStorageAccountContainer(t, credential, subscriptionID, externalResourceGroupName, storageAccountOneName, "test-container-a")
	storageAccountOneContainerB := CreateStorageAccountContainer(t, credential, subscriptionID, externalResourceGroupName, storageAccountOneName, "test-container-b")

	storageAccountTwoName := fmt.Sprintf("sa%sexternal2", strings.ToLower(uniqueId))
	storageAccountTwo := CreateStorageAccount(t, credential, subscriptionID, externalResourceGroupName, storageAccountTwoName, resourceGroupLocation)
	storageAccountTwoContainerA := CreateStorageAccountContainer(t, credential, subscriptionID, externalResourceGroupName, storageAccountTwoName, "test-container-a")

	externalResources := &TestBlobStorageBackupExternalResources{
		ResourceGroup:               resourceGroup,
		LogAnalyticsWorkspace:       logAnalyticsWorkspace,
		StorageAccountOne:           storageAccountOne,
		StorageAccountOneContainerA: storageAccountOneContainerA,
		StorageAccountOneContainerB: storageAccountOneContainerB,
		StorageAccountTwo:           storageAccountTwo,
		StorageAccountTwoContainerA: storageAccountTwoContainerA,
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

	uniqueId := random.UniqueId()
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", uniqueId)
	resourceGroupLocation := "uksouth"
	backupVaultName := fmt.Sprintf("bvault-nhsbackup-%s", uniqueId)

	externalResources := setupExternalResourcesForBlobStorageBackupTest(t, credential, environment.SubscriptionID, resourceGroupName, resourceGroupLocation, uniqueId)

	// A map of backups which we'll use to apply the TF module, and then validate the
	// policies have been created correctly
	blobStorageBackups := map[string]map[string]interface{}{
		"backup1": {
			"backup_name":                "blob1",
			"retention_period":           "P1D",
			"backup_intervals":           []string{"R/2024-01-01T00:00:00+00:00/P1D"},
			"storage_account_id":         *externalResources.StorageAccountOne.ID,
			"storage_account_containers": []string{*externalResources.StorageAccountOneContainerA.Name},
		},
		"backup2": {
			"backup_name":                "blob2",
			"retention_period":           "P1D",
			"backup_intervals":           []string{"R/2024-01-01T00:00:00+00:00/P1D"},
			"storage_account_id":         *externalResources.StorageAccountOne.ID,
			"storage_account_containers": []string{*externalResources.StorageAccountOneContainerB.Name},
		},
		"backup3": {
			"backup_name":                "blob3",
			"retention_period":           "P7D",
			"backup_intervals":           []string{"R/2024-01-01T00:00:00+00:00/P2D"},
			"storage_account_id":         *externalResources.StorageAccountTwo.ID,
			"storage_account_containers": []string{*externalResources.StorageAccountTwoContainerA.Name},
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
				"resource_group_name":        resourceGroupName,
				"resource_group_location":    resourceGroupLocation,
				"backup_vault_name":          backupVaultName,
				"log_analytics_workspace_id": *externalResources.LogAnalyticsWorkspace.ID,
				"blob_storage_backups":       blobStorageBackups,
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
		backupVault := GetBackupVault(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName)
		backupPolicies := GetBackupPolicies(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName)
		backupInstances := GetBackupInstances(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName)

		assert.Equal(t, len(blobStorageBackups), len(backupPolicies), "Expected to find %2 backup policies in vault", len(blobStorageBackups))
		assert.Equal(t, len(blobStorageBackups), len(backupInstances), "Expected to find %2 backup instances in vault", len(blobStorageBackups))

		for _, backup := range blobStorageBackups {
			backupName := backup["backup_name"].(string)
			retentionPeriod := backup["retention_period"].(string)
			backupIntervals := backup["backup_intervals"].([]string)
			storageAccountId := backup["storage_account_id"].(string)

			// Validate backup policy
			backupPolicyName := fmt.Sprintf("bkpol-blob-%s", backupName)
			backupPolicy := GetBackupPolicyForName(backupPolicies, backupPolicyName)
			assert.NotNil(t, backupPolicy, "Expected to find a backup policy called %s", backupPolicyName)

			// Validate retention period
			backupPolicyProperties := backupPolicy.Properties.(*armdataprotection.BackupPolicy)
			retentionRule := GetBackupPolicyRuleForName(backupPolicyProperties.PolicyRules, "Default").(*armdataprotection.AzureRetentionRule)
			deleteOption := retentionRule.Lifecycles[0].DeleteAfter.(*armdataprotection.AbsoluteDeleteOption)
			assert.Equal(t, retentionPeriod, *deleteOption.Duration, "Expected the backup policy retention period to be %s", retentionPeriod)

			// Validate backup intervals
			backupRule := GetBackupPolicyRuleForName(backupPolicyProperties.PolicyRules, "BackupIntervals").(*armdataprotection.AzureBackupRule)
			schedule := backupRule.Trigger.(*armdataprotection.ScheduleBasedTriggerContext).Schedule
			for index, interval := range schedule.RepeatingTimeIntervals {
				assert.Equal(t, backupIntervals[index], *interval, "Expected backup policy repeating interval %s to be %s", index, backupIntervals[index])
			}

			// Validate backup instance
			backupInstanceName := fmt.Sprintf("bkinst-blob-%s", backupName)
			backupInstance := GetBackupInstanceForName(backupInstances, backupInstanceName)
			assert.NotNil(t, backupInstance, "Expected to find a backup policy called %s", backupInstanceName)
			assert.Equal(t, storageAccountId, *backupInstance.Properties.DataSourceInfo.ResourceID, "Expected the backup instance source resource ID to be %s", storageAccountId)
			assert.Equal(t, *backupPolicy.ID, *backupInstance.Properties.PolicyInfo.PolicyID, "Expected the backup instance policy ID to be %s", backupPolicy.ID)

			// TODO: Validate storage containers here

			// Validate role assignment
			backupContributorRoleDefinition := GetRoleDefinition(t, credential, "Storage Account Backup Contributor")
			backupContributorRoleAssignment := GetRoleAssignment(t, credential, environment.SubscriptionID, *backupVault.Identity.PrincipalID, backupContributorRoleDefinition, storageAccountId)
			assert.NotNil(t, backupContributorRoleAssignment, "Expected to find role assignment %s for principal %s on scope %s", backupContributorRoleDefinition.Name, *backupVault.Identity.PrincipalID, storageAccountId)
		}
	})
}
