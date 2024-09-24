package e2e_tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

type ExternalResources struct {
	ResourceGroup     armresources.ResourceGroup
	StorageAccountOne armstorage.Account
	StorageAccountTwo armstorage.Account
}

/*
 * TestBasicDeployment tests the basic deployment of the infrastructure using Terraform.
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

	externalResources := setupExternalResources(t, credential, environment.SubscriptionID, vaultName, vaultLocation)

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

		destroyExternalResources(t, credential, environment.SubscriptionID, externalResources)
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

		validateBlobStorageBackups(t, credential, environment.SubscriptionID, vaultName, backupVault, backupPolicies, backupInstances, blobStorageBackups)
	})
}

/*
 * Validates the backup instances have been deployed correctly
 */
func validateBlobStorageBackups(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string,
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

		validateBackupPolicy(t, backupPolicy, retentionPeriod)
		validateBackupInstance(t, backupInstance, backupPolicy, storageAccountId)
		validateRoleAssignment(t, credential, subscriptionID, *backupVault.Identity.PrincipalID, "Storage Account Backup Contributor", storageAccountId)
	}
}

/*
 * Validates a role assignment.
 */
func validateRoleAssignment(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string,
	principalId string, roleDefinitionName string, scope string) {
	client, err := armauthorization.NewRoleAssignmentsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create role assignments client: %v", err)

	// List role assignments for the given scope
	filter := fmt.Sprintf("principalId eq '%s'", principalId)
	pager := client.NewListForScopePager(scope, &armauthorization.RoleAssignmentsClientListForScopeOptions{Filter: &filter})

	// Check if the role definition is among the assigned roles
	found := false
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		assert.NoError(t, err, "Failed to list role assignments")

		for _, roleAssignment := range page.RoleAssignmentListResult.Value {
			roleDefinitionsClient, err := armauthorization.NewRoleDefinitionsClient(credential, nil)
			assert.NoError(t, err, "Failed to create role definitions client: %v", err)

			roleDefinition, err := roleDefinitionsClient.Get(context.Background(), scope, *roleAssignment.Properties.RoleDefinitionID, nil)
			assert.NoError(t, err, "Failed to get role definition")

			if *roleDefinition.Properties.RoleName == roleDefinitionName {
				found = true
				break
			}
		}
	}

	assert.True(t, found, "Expected to find role assignment %s for principal %s on scope %s", roleDefinitionName, principalId, scope)
}

/*
 * Validates a backup instance
 */
func validateBackupInstance(t *testing.T, instance *armdataprotection.BackupInstanceResource, policy *armdataprotection.BaseBackupPolicyResource, expectedStorageAccountId string) {
	assert.Equal(t, expectedStorageAccountId, instance.Properties.DataSourceInfo.ResourceID, "Expected the backup instance source resource ID to be %s", expectedStorageAccountId)
	assert.Equal(t, policy.ID, instance.Properties.PolicyInfo.PolicyID, "Expected the backup instance source resource ID to be %s", expectedStorageAccountId)
}

/*
 * Validates a backup policy
 */
func validateBackupPolicy(t *testing.T, policy *armdataprotection.BaseBackupPolicyResource, expectedRetentionPeriod string) {
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
func setupExternalResources(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, vault_name string, vault_location string) *ExternalResources {
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s-external", vault_name)
	resourceGroup := CreateResourceGroup(t, subscriptionID, credential, resourceGroupName, vault_location)

	storageAccountOneName := fmt.Sprintf("sa%sexternal1", strings.ToLower(vault_name))
	storageAccountOne := CreateStorageAccount(t, credential, subscriptionID, resourceGroupName, storageAccountOneName, vault_location)

	storageAccountTwoName := fmt.Sprintf("sa%sexternal2", strings.ToLower(vault_name))
	storageAccountTwo := CreateStorageAccount(t, credential, subscriptionID, resourceGroupName, storageAccountTwoName, vault_location)

	externalResources := &ExternalResources{
		ResourceGroup:     resourceGroup,
		StorageAccountOne: storageAccountOne,
		StorageAccountTwo: storageAccountTwo,
	}

	return externalResources
}

/*
 * Destroys the external resources.
 */
func destroyExternalResources(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, externalResources *ExternalResources) {
	DeleteResourceGroup(t, credential, subscriptionID, *externalResources.ResourceGroup.Name)
}
