package e2e_tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

type TestManagedDiskBackupExternalResources struct {
	ResourceGroup  armresources.ResourceGroup
	ManagedDiskOne armcompute.Disk
	ManagedDiskTwo armcompute.Disk
	LogAnalyticsWorkspace armoperationalinsights.Workspace
}

/*
 * Creates resources which are "external" to the az-backup module, and models
 * what would be backed up in a real scenario.
 */
func setupExternalResourcesForManagedDiskBackupTest(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, resourceGroupLocation string, uniqueId string) *TestManagedDiskBackupExternalResources {
	externalResourceGroupName := fmt.Sprintf("%s-external", resourceGroupName)
	resourceGroup := CreateResourceGroup(t, credential, subscriptionID, externalResourceGroupName, resourceGroupLocation)

	managedDiskOneName := fmt.Sprintf("disk-%s-external-1", strings.ToLower(uniqueId))
	managedDiskOne := CreateManagedDisk(t, credential, subscriptionID, externalResourceGroupName, managedDiskOneName, resourceGroupLocation, int32(1))

	managedDiskTwoName := fmt.Sprintf("disk-%s-external-2", strings.ToLower(uniqueId))
	managedDiskTwo := CreateManagedDisk(t, credential, subscriptionID, externalResourceGroupName, managedDiskTwoName, resourceGroupLocation, int32(1))

	logAnalyticsWorkspaceName := fmt.Sprintf("law-%s-external", strings.ToLower(uniqueId))
	logAnalyticsWorkspace := CreateLogAnalyticsWorkspace(t, credential, subscriptionID, externalResourceGroupName, logAnalyticsWorkspaceName, resourceGroupLocation)

	externalResources := &TestManagedDiskBackupExternalResources{
		ResourceGroup:         resourceGroup,
		ManagedDiskOne:        managedDiskOne,
		ManagedDiskTwo:        managedDiskTwo,
		LogAnalyticsWorkspace: logAnalyticsWorkspace,
	}

	return externalResources
}

/*
 * TestManagedDiskBackup tests the deployment of a backup vault and backup policies for managed disks.
 */
func TestManagedDiskBackup(t *testing.T) {
	t.Parallel()

	environment := GetEnvironmentConfiguration(t)
	credential := GetAzureCredential(t, environment)

	uniqueId := random.UniqueId()
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", uniqueId)
	resourceGroupLocation := "uksouth"
	backupVaultName := fmt.Sprintf("bvault-nhsbackup-%s", uniqueId)

	externalResources := setupExternalResourcesForManagedDiskBackupTest(t, credential, environment.SubscriptionID, resourceGroupName, resourceGroupLocation, uniqueId)

	// A map of backups which we'll use to apply the TF module, and then validate the
	// policies have been created correctly
	managedDiskBackups := map[string]map[string]interface{}{
		"backup1": {
			"backup_name":      "disk1",
			"retention_period": "P1D",
			"backup_intervals": []string{"R/2024-01-01T00:00:00+00:00/P1D"},
			"managed_disk_id":  *externalResources.ManagedDiskOne.ID,
			"managed_disk_resource_group": map[string]interface{}{
				"id":   *externalResources.ResourceGroup.ID,
				"name": *externalResources.ResourceGroup.Name,
			},
		},
		"backup2": {
			"backup_name":      "disk2",
			"retention_period": "P7D",
			"backup_intervals": []string{"R/2024-01-01T00:00:00+00:00/P2D"},
			"managed_disk_id":  *externalResources.ManagedDiskTwo.ID,
			"managed_disk_resource_group": map[string]interface{}{
				"id":   *externalResources.ResourceGroup.ID,
				"name": *externalResources.ResourceGroup.Name,
			},
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
				"managed_disk_backups":       managedDiskBackups,
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
		backupVault := GetBackupVault(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName)
		backupPolicies := GetBackupPolicies(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName)
		backupInstances := GetBackupInstances(t, credential, environment.SubscriptionID, resourceGroupName, backupVaultName)

		assert.Equal(t, len(managedDiskBackups), len(backupPolicies), "Expected to find %2 backup policies in vault", len(managedDiskBackups))
		assert.Equal(t, len(managedDiskBackups), len(backupInstances), "Expected to find %2 backup instances in vault", len(managedDiskBackups))

		for _, backup := range managedDiskBackups {
			backupName := backup["backup_name"].(string)
			retentionPeriod := backup["retention_period"].(string)
			backupIntervals := backup["backup_intervals"].([]string)
			managedDiskId := backup["managed_disk_id"].(string)
			managedDiskResourceGroup := backup["managed_disk_resource_group"].(map[string]interface{})
			managedDiskResourceGroupId := managedDiskResourceGroup["id"].(string)

			// Validate backup policy
			backupPolicyName := fmt.Sprintf("bkpol-disk-%s", backupName)
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
			backupInstanceName := fmt.Sprintf("bkinst-disk-%s", backupName)
			backupInstance := GetBackupInstanceForName(backupInstances, backupInstanceName)
			assert.NotNil(t, backupInstance, "Expected to find a backup policy called %s", backupInstanceName)
			assert.Equal(t, managedDiskId, *backupInstance.Properties.DataSourceInfo.ResourceID, "Expected the backup instance source resource ID to be %s", managedDiskId)
			assert.Equal(t, *backupPolicy.ID, *backupInstance.Properties.PolicyInfo.PolicyID, "Expected the backup instance policy ID to be %s", backupPolicy.ID)

			// Validate role assignments
			snapshotContributorRoleDefinition := GetRoleDefinition(t, credential, "Disk Snapshot Contributor")
			snapshotContributorRoleAssignment := GetRoleAssignment(t, credential, environment.SubscriptionID, *backupVault.Identity.PrincipalID, snapshotContributorRoleDefinition, managedDiskResourceGroupId)
			assert.NotNil(t, snapshotContributorRoleAssignment, "Expected to find role assignment %s for principal %s on scope %s", snapshotContributorRoleDefinition.Name, *backupVault.Identity.PrincipalID, managedDiskResourceGroupId)

			backupReaderRoleDefinition := GetRoleDefinition(t, credential, "Disk Backup Reader")
			backupReaderRoleAssignment := GetRoleAssignment(t, credential, environment.SubscriptionID, *backupVault.Identity.PrincipalID, backupReaderRoleDefinition, managedDiskId)
			assert.NotNil(t, backupReaderRoleAssignment, "Expected to find role assignment %s for principal %s on scope %s", backupReaderRoleDefinition.Name, *backupVault.Identity.PrincipalID, managedDiskId)
		}
	})
}
