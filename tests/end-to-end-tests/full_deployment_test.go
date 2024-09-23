package e2e_tests

import (
	"context"
	"fmt"
	"os"
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
 * TestFullDeployment tests the full deployment of the infrastructure using Terraform.
 */
func TestFullDeployment(t *testing.T) {
	t.Parallel()

	terraformFolder := test_structure.CopyTerraformFolderToTemp(t, "../../infrastructure", "")
	terraformStateResourceGroup := os.Getenv("TF_STATE_RESOURCE_GROUP")
	terraformStateStorageAccount := os.Getenv("TF_STATE_STORAGE_ACCOUNT")
	terraformStateContainer := os.Getenv("TF_STATE_STORAGE_CONTAINER")

	if terraformStateResourceGroup == "" || terraformStateStorageAccount == "" || terraformStateContainer == "" {
		t.Fatalf("One or more required environment variables (TF_STATE_RESOURCE_GROUP, TF_STATE_STORAGE_ACCOUNT, TF_STATE_STORAGE_CONTAINER) are not set.")
	}

	vaultName := random.UniqueId()
	vaultLocation := "uksouth"
	vaultRedundancy := "LocallyRedundant"

	// Setup stage
	// ...

	test_structure.RunTestStage(t, "setup", func() {
		terraformOptions := &terraform.Options{
			TerraformDir: terraformFolder,

			// Variables to pass to our Terraform code using -var options
			Vars: map[string]interface{}{
				"vault_name":       vaultName,
				"vault_location":   vaultLocation,
				"vault_redundancy": vaultRedundancy,
			},

			BackendConfig: map[string]interface{}{
				"resource_group_name":  terraformStateResourceGroup,
				"storage_account_name": terraformStateStorageAccount,
				"container_name":       terraformStateContainer,
				"key":                  vaultName + ".tfstate",
			},
		}

		// Save options for later test stages
		test_structure.SaveTerraformOptions(t, terraformFolder, terraformOptions)

		terraform.InitAndApply(t, terraformOptions)
	})

	// Validate stage
	// ...

	test_structure.RunTestStage(t, "validate", func() {
		resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", vaultName)
		fullVaultName := fmt.Sprintf("bvault-%s", vaultName)

		// Get credentials from environment variables
		tenantID := os.Getenv("ARM_TENANT_ID")
		subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
		clientID := os.Getenv("ARM_CLIENT_ID")
		clientSecret := os.Getenv("ARM_CLIENT_SECRET")

		if tenantID == "" || subscriptionID == "" || clientID == "" || clientSecret == "" {
			t.Fatalf("One or more required environment variables (ARM_TENANT_ID, ARM_SUBSCRIPTION_ID, ARM_CLIENT_ID, ARM_CLIENT_SECRET) are not set.")
		}

		// Create a credential to authenticate with Azure Resource Manager
		cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
		assert.NoError(t, err, "Failed to obtain a credential: %v", err)

		ValidateResourceGroup(t, subscriptionID, cred, resourceGroupName, vaultLocation)
		ValidateBackupVault(t, subscriptionID, cred, resourceGroupName, fullVaultName, vaultLocation)
		ValidateBackupPolicies(t, subscriptionID, cred, resourceGroupName, fullVaultName, vaultName)
	})

	// Teardown stage
	// ...

	test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformFolder)

		terraform.Destroy(t, terraformOptions)
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

/*
 * Validates the backup policies have been deployed correctly
 */
func ValidateBackupPolicies(t *testing.T, subscriptionID string, cred *azidentity.ClientSecretCredential, resourceGroupName string, fullVaultName string, vaultName string) {
	// Create a client to interact with Data Protection vault backup policies
	client, err := armdataprotection.NewBackupPoliciesClient(subscriptionID, cred, nil)
	assert.NoError(t, err, "Failed to create data protection client: %v", err)

	policyPager := client.NewListPager(resourceGroupName, fullVaultName, nil)

	// Fetch all backup policies from the vault
	var policies []*armdataprotection.BaseBackupPolicyResource

	for policyPager.More() {
		page, err := policyPager.NextPage(context.Background())
		assert.NoError(t, err, "Failed to get backup policies: %v", err)

		policies = append(policies, page.Value...)
	}

	// Validate the policies
	if len(policies) == 0 {
		assert.Fail(t, "Expected to find at least one backup policy in vault %s", fullVaultName)
	} else {
		assert.Equal(t, len(policies), 2, "Expected to find two backup policies in vault %s", fullVaultName)

		ValidateManagedDiskPolicy(t, policies, vaultName)
		ValidateBlobStoragePolicy(t, policies, vaultName)
	}
}

/*
 * Validates the blob storage backup policy
 */
func ValidateBlobStoragePolicy(t *testing.T, policies []*armdataprotection.BaseBackupPolicyResource, vaultName string) {
	blobStoragePolicyName := fmt.Sprintf("bkpol-%s-blobstorage", vaultName)
	blobStoragePolicy := GetBackupPolicyForName(policies, blobStoragePolicyName)
	assert.NotNil(t, blobStoragePolicy, "Expected to find a blob storage backup policy called %s", blobStoragePolicyName)

	blobStoragePolicyProperties, ok := blobStoragePolicy.Properties.(*armdataprotection.BackupPolicy)
	assert.True(t, ok, "Failed to cast blob storage policy properties to BackupPolicy")

	// Validate the retention policy
	retentionPeriodPolicyRule := GetBackupPolicyRuleForName(blobStoragePolicyProperties.PolicyRules, "Default")
	assert.NotNil(t, retentionPeriodPolicyRule, "Expected to find a policy rule called Default in the blob storage backup policies")

	azureRetentionRule, ok := retentionPeriodPolicyRule.(*armdataprotection.AzureRetentionRule)
	assert.True(t, ok, "Failed to cast retention period policy rule to AzureRetentionRule")

	deleteOption, ok := azureRetentionRule.Lifecycles[0].DeleteAfter.(*armdataprotection.AbsoluteDeleteOption)
	assert.True(t, ok, "Failed to cast delete option to AbsoluteDeleteOption")

	assert.Equal(t, "P7D", *deleteOption.Duration, "Expected the blob storage retention period to be P7D")
}

/*
 * Validates the managed disk backup policy
 */
func ValidateManagedDiskPolicy(t *testing.T, policies []*armdataprotection.BaseBackupPolicyResource, vaultName string) {
	managedDiskPolicyName := fmt.Sprintf("bkpol-%s-manageddisk", vaultName)
	managedDiskPolicy := GetBackupPolicyForName(policies, managedDiskPolicyName)
	assert.NotNil(t, managedDiskPolicy, "Expected to find a managed disk backup policy called %s", managedDiskPolicyName)

	managedDiskPolicyProperties, ok := managedDiskPolicy.Properties.(*armdataprotection.BackupPolicy)
	assert.True(t, ok, "Failed to cast managed disk policy properties to BackupPolicy")

	// Validate the repeating time intervals
	backupIntervalsPolicyRule := GetBackupPolicyRuleForName(managedDiskPolicyProperties.PolicyRules, "BackupIntervals")
	assert.NotNil(t, backupIntervalsPolicyRule, "Expected to find a policy rule called BackupIntervals in the managed disk backup policies")

	azureBackupRule, ok := backupIntervalsPolicyRule.(*armdataprotection.AzureBackupRule)
	assert.True(t, ok, "Failed to cast backup intervals policy rule to AzureBackupRule")

	trigger, ok := azureBackupRule.Trigger.(*armdataprotection.ScheduleBasedTriggerContext)
	assert.True(t, ok, "Failed to cast azure backup rule trigger to ScheduleBasedTriggerContext")

	assert.Equal(t, "R/2024-01-01T00:00:00+00:00/P1D", *trigger.Schedule.RepeatingTimeIntervals[0],
		"Expected the managed disk backup policy repeating time intervals to be R/2024-01-01T00:00:00+00:00/P1D")

	// Validate the retention policy
	retentionPeriodPolicyRule := GetBackupPolicyRuleForName(managedDiskPolicyProperties.PolicyRules, "Default")
	assert.NotNil(t, retentionPeriodPolicyRule, "Expected to find a policy rule called Default in the managed disk backup policies")

	azureRetentionRule, ok := retentionPeriodPolicyRule.(*armdataprotection.AzureRetentionRule)
	assert.True(t, ok, "Failed to cast retention period policy rule to AzureRetentionRule")

	deleteOption, ok := azureRetentionRule.Lifecycles[0].DeleteAfter.(*armdataprotection.AbsoluteDeleteOption)
	assert.True(t, ok, "Failed to cast delete option to AbsoluteDeleteOption")

	assert.Equal(t, "P7D", *deleteOption.Duration, "Expected the managed disk retention period to be P7D")
}

/*
 * Gets a backup policy from the provided list for the provided name
 */
func GetBackupPolicyForName(policies []*armdataprotection.BaseBackupPolicyResource, name string) *armdataprotection.BaseBackupPolicyResource {
	for _, policy := range policies {
		if *policy.Name == name {
			return policy
		}
	}

	return nil
}

/*
 * Gets a backup policy rules from the provided list for the provided name
 */
func GetBackupPolicyRuleForName(policyRules []armdataprotection.BasePolicyRuleClassification, name string) armdataprotection.BasePolicyRuleClassification {
	for _, policyRule := range policyRules {
		if *policyRule.GetBasePolicyRule().Name == name {
			return policyRule
		}
	}

	return nil
}
