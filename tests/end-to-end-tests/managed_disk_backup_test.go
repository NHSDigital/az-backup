package e2e_tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

type TestManagedDiskBackupExternalResources struct {
	ResourceGroup armresources.ResourceGroup
	ManagedDisk1  armcompute.Disk
	ManagedDisk2  armcompute.Disk
}

/*
 * TestManagedDiskBackup tests the deployment of a backup vault and backup policies for blob storage accounts.
 */
func TestManagedDiskBackup(t *testing.T) {
	t.Parallel()

	environment := GetEnvironmentConfiguration(t)
	credential := GetAzureCredential(t, environment)

	vaultName := random.UniqueId()
	vaultLocation := "uksouth"
	vaultRedundancy := "LocallyRedundant"
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", vaultName)
	backupVaultName := fmt.Sprintf("bvault-%s", vaultName)

	externalResources := testManagedDiskBackupSetupExternalResources(t, credential, environment.SubscriptionID, vaultName, vaultLocation)

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

	// Teardown stage - deferred so it runs after the other test stages
	// regardless of whether they succeed or fail.
	// ...

	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, environment.TerraformFolder)

		terraform.Destroy(t, terraformOptions)

		testManagedDiskBackupDestroyExternalResources(t, credential, environment.SubscriptionID, externalResources)
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

		testManagedDiskBackupValidateBackups(t, credential, environment.SubscriptionID, vaultName, backupVault, backupPolicies, backupInstances, blobStorageBackups)
	})
}

/*
 * Validates the backup instances have been deployed correctly
 */
func testManagedDiskBackupValidateBackups(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string,
	vaultName string, backupVault armdataprotection.BackupVaultResource, policies []*armdataprotection.BaseBackupPolicyResource,
	instances []*armdataprotection.BackupInstanceResource, blobStorageBackups map[string]map[string]interface{}) {
	assert.Equal(t, len(blobStorageBackups), len(instances), "Expected to find %2 backup instances in vault", len(blobStorageBackups))

	for _, backup := range blobStorageBackups {
		backupName := backup["backup_name"].(string)
		retentionPeriod := backup["retention_period"].(string)
		storageAccountId := backup["storage_account_id"].(string)

		backupPolicyName := fmt.Sprintf("bkpol-%s-blobstorage-%s", vaultName, backupName)
		backupPolicy := GetBackupPolicyForName(policies, backupPolicyName)
		assert.NotNil(t, backupPolicy, "Expected to find a backup policy called %s", backupPolicyName)

		backupInstanceName := fmt.Sprintf("bkinst-%s-blobstorage-%s", vaultName, backupName)
		backupInstance := GetBackupInstanceForName(instances, backupInstanceName)
		assert.NotNil(t, backupInstance, "Expected to find a backup policy called %s", backupInstanceName)

		roleDefinition := GetRoleDefinition(t, credential, "Storage Account Backup Contributor")
		assert.NotNil(t, roleDefinition, "Expected to find a role definition called %s", "Storage Account Backup Contributor")

		testManagedDiskBackupValidateBackupPolicy(t, backupPolicy, retentionPeriod)
		testManagedDiskBackupValidateBackupInstance(t, backupInstance, backupPolicy, storageAccountId)
		testManagedDiskBackupValidateRoleAssignment(t, credential, subscriptionID, *backupVault.Identity.PrincipalID, roleDefinition, storageAccountId)
	}
}

/*
 * Validates a role assignment.
 */
func testManagedDiskBackupValidateRoleAssignment(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string,
	principalId string, roleDefinition *armauthorization.RoleDefinition, storageAccountId string) {
	roleAssignmentsClient, err := armauthorization.NewRoleAssignmentsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create role assignments client: %v", err)

	// List role assignments for the given scope
	filter := fmt.Sprintf("principalId eq '%s'", principalId)
	pager := roleAssignmentsClient.NewListForScopePager(storageAccountId, &armauthorization.RoleAssignmentsClientListForScopeOptions{Filter: &filter})

	// Check if the role definition is among the assigned roles
	found := false
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		assert.NoError(t, err, "Failed to list role assignments")

		// Check if the role definition is among the assigned roles
		for _, roleAssignment := range page.RoleAssignmentListResult.Value {
			if strings.Contains(*roleAssignment.Properties.RoleDefinitionID, *roleDefinition.ID) {
				found = true
			}
		}
	}

	assert.True(t, found, "Expected to find role assignment %s for principal %s on scope %s", roleDefinition.Name, principalId, storageAccountId)
}

/*
 * Validates a backup instance
 */
func testManagedDiskBackupValidateBackupInstance(t *testing.T, instance *armdataprotection.BackupInstanceResource, policy *armdataprotection.BaseBackupPolicyResource, expectedStorageAccountId string) {
	assert.Equal(t, expectedStorageAccountId, *instance.Properties.DataSourceInfo.ResourceID, "Expected the backup instance source resource ID to be %s", expectedStorageAccountId)
	assert.Equal(t, *policy.ID, *instance.Properties.PolicyInfo.PolicyID, "Expected the backup instance policy ID to be %s", policy.ID)
}

/*
 * Validates a backup policy
 */
func testManagedDiskBackupValidateBackupPolicy(t *testing.T, policy *armdataprotection.BaseBackupPolicyResource, expectedRetentionPeriod string) {
	blobStoragePolicyProperties := policy.Properties.(*armdataprotection.BackupPolicy)
	retentionPeriodPolicyRule := GetBackupPolicyRuleForName(blobStoragePolicyProperties.PolicyRules, "Default")
	assert.NotNil(t, retentionPeriodPolicyRule, "Expected to find a policy rule called Default in the backup policies")

	azureRetentionRule := retentionPeriodPolicyRule.(*armdataprotection.AzureRetentionRule)
	deleteOption := azureRetentionRule.Lifecycles[0].DeleteAfter.(*armdataprotection.AbsoluteDeleteOption)
	assert.Equal(t, expectedRetentionPeriod, *deleteOption.Duration, "Expected the backup policy retention period to be %s", expectedRetentionPeriod)
}

/*
 * Creates resources which are "external" to the az-backup module, and will be backed up in a real scenario.
 */
func testManagedDiskBackupSetupExternalResources(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, vault_name string, vault_location string) *TestBlobStorageBackupExternalResources {
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s-external", vault_name)
	resourceGroup := CreateResourceGroup(t, subscriptionID, credential, resourceGroupName, vault_location)

	managedDiskOneName := fmt.Sprintf("sa%sexternal1", strings.ToLower(vault_name))
	managedDiskOne := CreateStorageAccount(t, credential, subscriptionID, resourceGroupName, managedDiskOneName, vault_location)

	managedDiskTwoName := fmt.Sprintf("sa%sexternal2", strings.ToLower(vault_name))
	managedDiskTwo := CreateStorageAccount(t, credential, subscriptionID, resourceGroupName, managedDiskTwoName, vault_location)

	externalResources := &TestManagedDiskBackupExternalResources{
		ResourceGroup:     resourceGroup,
		StorageAccountOne: managedDiskOne,
		StorageAccountTwo: managedDiskTwo,
	}

	return externalResources
}

/*
 * Destroys the external resources.
 */
func testManagedDiskBackupDestroyExternalResources(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, externalResources *TestBlobStorageBackupExternalResources) {
	DeleteResourceGroup(t, credential, subscriptionID, *externalResources.ResourceGroup.Name)
}
