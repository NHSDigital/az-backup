package e2e_tests

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

type Config struct {
	TerraformFolder              string
	TenantID                     string
	SubscriptionID               string
	ClientID                     string
	ClientSecret                 string
	TerraformStateResourceGroup  string
	TerraformStateStorageAccount string
	TerraformStateContainer      string
}

/*
 * GetEnvironmentConfiguration gets the configuration for the test environment.
 */
func GetEnvironmentConfiguration(t *testing.T) *Config {
	terraformFolder := test_structure.CopyTerraformFolderToTemp(t, "../../infrastructure", "")

	tenantID := os.Getenv("ARM_TENANT_ID")
	if tenantID == "" {
		t.Fatalf("ARM_TENANT_ID must be set")
	}

	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		t.Fatalf("ARM_SUBSCRIPTION_ID must be set")
	}

	clientID := os.Getenv("ARM_CLIENT_ID")
	if clientID == "" {
		t.Fatalf("ARM_CLIENT_ID must be set")
	}

	clientSecret := os.Getenv("ARM_CLIENT_SECRET")
	if clientSecret == "" {
		t.Fatalf("ARM_CLIENT_SECRET must be set")
	}

	terraformStateResourceGroup := os.Getenv("TF_STATE_RESOURCE_GROUP")
	if terraformStateResourceGroup == "" {
		t.Fatalf("TF_STATE_RESOURCE_GROUP must be set")
	}

	terraformStateStorageAccount := os.Getenv("TF_STATE_STORAGE_ACCOUNT")
	if terraformStateStorageAccount == "" {
		t.Fatalf("TF_STATE_STORAGE_ACCOUNT must be set")
	}

	terraformStateContainer := os.Getenv("TF_STATE_STORAGE_CONTAINER")
	if terraformStateContainer == "" {
		t.Fatalf("TF_STATE_STORAGE_CONTAINER must be set")
	}

	config := &Config{
		TerraformFolder:              terraformFolder,
		TenantID:                     tenantID,
		SubscriptionID:               subscriptionID,
		ClientID:                     clientID,
		ClientSecret:                 clientSecret,
		TerraformStateResourceGroup:  terraformStateResourceGroup,
		TerraformStateStorageAccount: terraformStateStorageAccount,
		TerraformStateContainer:      terraformStateContainer,
	}

	return config
}

/*
 * Gets a credential for authenticating with Azure Resource Manager.
 */
func GetAzureCredential(t *testing.T, environment *Config) *azidentity.ClientSecretCredential {
	credential, err := azidentity.NewClientSecretCredential(environment.TenantID, environment.ClientID, environment.ClientSecret, nil)
	assert.NoError(t, err, "Failed to obtain a credential: %v", err)

	return credential
}

/*
 * Gets the backup vault for the provided name.
 */
func GetBackupVault(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, backupVaultName string) armdataprotection.BackupVaultResource {
	client, err := armdataprotection.NewBackupVaultsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create data protection client: %v", err)

	// Get the backup vault
	resp, err := client.Get(context.Background(), resourceGroupName, backupVaultName, nil)
	assert.NoError(t, err, "Failed to get backup vault: %v", err)

	return resp.BackupVaultResource
}

/*
 * Gets the backup policies for the provided backup vault.
 */
func GetBackupPolicies(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, backupVaultName string) []*armdataprotection.BaseBackupPolicyResource {
	client, err := armdataprotection.NewBackupPoliciesClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create data protection client: %v", err)

	policyPager := client.NewListPager(resourceGroupName, backupVaultName, nil)

	var policies []*armdataprotection.BaseBackupPolicyResource

	for policyPager.More() {
		page, err := policyPager.NextPage(context.Background())
		assert.NoError(t, err, "Failed to get backup policies: %v", err)

		policies = append(policies, page.Value...)
	}

	return policies
}

/*
 * Gets the backup instances for the provided backup vault.
 */
func GetBackupInstances(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, backupVaultName string) []*armdataprotection.BackupInstanceResource {
	client, err := armdataprotection.NewBackupInstancesClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create data protection client: %v", err)

	policyPager := client.NewListPager(resourceGroupName, backupVaultName, nil)

	var instances []*armdataprotection.BackupInstanceResource

	for policyPager.More() {
		page, err := policyPager.NextPage(context.Background())
		assert.NoError(t, err, "Failed to get backup policies: %v", err)

		instances = append(instances, page.Value...)
	}

	return instances
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

/*
 * Gets a backup policy from the provided list for the provided name
 */
func GetBackupInstanceForName(instances []*armdataprotection.BackupInstanceResource, name string) *armdataprotection.BackupInstanceResource {
	for _, instance := range instances {
		if *instance.Name == name {
			return instance
		}
	}

	return nil
}

/*
 * Creates a resource group that can be used for testing purposes.
 */
func CreateResourceGroup(t *testing.T, subscriptionID string, credential *azidentity.ClientSecretCredential, resourceGroupName string, resourceGroupLocation string) armresources.ResourceGroup {
	client, err := armresources.NewResourceGroupsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create resource group client: %v", err)

	t.Logf("Creating resource group %s in location %s", resourceGroupName, resourceGroupLocation)

	resp, err := client.CreateOrUpdate(
		context.Background(),
		resourceGroupName,
		armresources.ResourceGroup{
			Location: &resourceGroupLocation,
		},
		nil,
	)
	assert.NoError(t, err, "Failed to create resource group: %v", err)

	t.Logf("Resource group %s created successfully", resourceGroupName)

	return resp.ResourceGroup
}

/*
 * Deletes a resource group.
 */
func DeleteResourceGroup(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string) {
	client, err := armresources.NewResourceGroupsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create resource group client: %v", err)

	t.Logf("Deleting resource group %s", resourceGroupName)

	pollerResp, err := client.BeginDelete(context.Background(), resourceGroupName, nil)
	assert.NoError(t, err, "Failed to delete resource group: %v", err)

	// Wait for the creation to complete
	_, err = pollerResp.PollUntilDone(context.Background(), nil)
	assert.NoError(t, err, "Failed to create storage account: %v", err)

	t.Logf("Resource group %s deleted successfully", resourceGroupName)
}

/*
 * Creates a storage account that can be used for testing purposes.
 */
func CreateStorageAccount(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string,
	resourceGroupName string, storageAccountName string, storageAccountLocation string) armstorage.Account {
	client, err := armstorage.NewAccountsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create storage account client: %v", err)

	t.Logf("Creating storage account %s in location %s", storageAccountName, storageAccountLocation)

	pollerResp, err := client.BeginCreate(
		context.Background(),
		resourceGroupName,
		storageAccountName,
		armstorage.AccountCreateParameters{
			SKU: &armstorage.SKU{
				Name: to.Ptr(armstorage.SKUNameStandardLRS),
			},
			Kind:     to.Ptr(armstorage.KindStorageV2),
			Location: &storageAccountLocation,
		},
		nil,
	)
	assert.NoError(t, err, "Failed to begin creating storage account: %v", err)

	// Wait for the creation to complete
	resp, err := pollerResp.PollUntilDone(context.Background(), nil)
	assert.NoError(t, err, "Failed to create storage account: %v", err)

	t.Logf("Storage account %s created successfully", storageAccountName)

	return resp.Account
}
