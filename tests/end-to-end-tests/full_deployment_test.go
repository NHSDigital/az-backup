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

func TestFullDeployment(t *testing.T) {
	t.Parallel()

	terraformFolder := "../../infrastructure"

	vaultName := random.UniqueId()
	vaultLocation := "uksouth"
	vaultRedundancy := "LocallyRedundant"

	// Setup stage
	test_structure.RunTestStage(t, "setup", func() {
		terraformOptions := &terraform.Options{
			TerraformDir: terraformFolder,

			// Variables to pass to our Terraform code using -var options
			Vars: map[string]interface{}{
				"vault_name":       vaultName,
				"vault_location":   vaultLocation,
				"vault_redundancy": vaultRedundancy,
			},
		}

		// Save options for later test stages
		test_structure.SaveTerraformOptions(t, terraformFolder, terraformOptions)

		terraform.InitAndApply(t, terraformOptions)
	})

	// Validate stage
	test_structure.RunTestStage(t, "validate", func() {
		resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", vaultName)
		fullVaultName := fmt.Sprintf("bvault-%s", vaultName)

		// Get credentials from environment variables
		tenantID := os.Getenv("ARM_TENANT_ID")
		subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
		clientID := os.Getenv("ARM_CLIENT_ID")
		clientSecret := os.Getenv("ARM_CLIENT_SECRET")

		// Validate that the required environment variables are set
		if tenantID == "" || subscriptionID == "" || clientID == "" || clientSecret == "" {
			t.Fatalf("One or more required environment variables (ARM_TENANT_ID, ARM_SUBSCRIPTION_ID, ARM_CLIENT_ID, ARM_CLIENT_SECRET) are not set.")
		}

		// Create a credential to authenticate with Azure Resource Manager
		cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
		assert.NoError(t, err, "Failed to obtain a credential: %v", err)

		// Check the resource group was created
		ValidateResourceGroup(t, subscriptionID, cred, resourceGroupName)

		// Check the backup vault was created
		ValidateBackupVault(t, subscriptionID, cred, resourceGroupName, fullVaultName)

		// Check the expected policies were created
		ValidateBackupPolicies(t, subscriptionID, cred, resourceGroupName, fullVaultName, vaultName)
	})

	// Teardown stage
	test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformFolder)

		terraform.Destroy(t, terraformOptions)
	})
}

func ValidateResourceGroup(t *testing.T, subscriptionID string, cred *azidentity.ClientSecretCredential, resourceGroupName string) {
	// Create a new resource groups client
	client, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	assert.NoError(t, err, "Failed to create resource group client: %v", err)
	assert.NoError(t, err)

	// Get the resource group
	resp, err := client.Get(context.Background(), resourceGroupName, nil)
	assert.NoError(t, err, "Failed to get resource group: %v", err)

	// Validate the resource group
	assert.NotNil(t, resp.ResourceGroup)
}

func ValidateBackupVault(t *testing.T, subscriptionID string, cred *azidentity.ClientSecretCredential, resourceGroupName string, vaultName string) {
	// Create a new Data Protection Backup Vaults client
	client, err := armdataprotection.NewBackupVaultsClient(subscriptionID, cred, nil)
	assert.NoError(t, err, "Failed to create data protection client: %v", err)

	// Get the backup vault
	resp, err := client.Get(context.Background(), resourceGroupName, vaultName, nil)
	assert.NoError(t, err, "Failed to get backup vault: %v", err)

	// Validate the backup vault exists
	assert.NotNil(t, resp.BackupVaultResource, "Backup vault does not exist")
	assert.Equal(t, *resp.BackupVaultResource.Name, vaultName, "Backup vault name does not match")
}

func ValidateBackupPolicies(t *testing.T, subscriptionID string, cred *azidentity.ClientSecretCredential, resourceGroupName string, fullVaultName string, vaultName string) {
	ctx := context.Background()

	// Create a client to interact with Data Protection vault backup policies
	client, err := armdataprotection.NewBackupPoliciesClient(subscriptionID, cred, nil)
	assert.NoError(t, err, "Failed to create data protection client: %v", err)

	policyPager := client.NewListPager(resourceGroupName, fullVaultName, nil)

	// Fetch all backup policies from the vault
	var policies []*armdataprotection.BaseBackupPolicyResource

	for policyPager.More() {
		page, err := policyPager.NextPage(ctx)
		assert.NoError(t, err, "Failed to get backup policies: %v", err)

		policies = append(policies, page.Value...)
	}

	// Validate the policies
	if len(policies) == 0 {
		assert.Fail(t, "Expected to find at least one backup policy in vault %s", fullVaultName)
	} else {
		assert.Equal(t, len(policies), 2, "Expected to find two backup policies in vault %s", fullVaultName)

		managedDiskPolicyName := fmt.Sprintf("bkpol-%s-manageddisk", vaultName)
		managedDiskPolicyExists := BackupPolicyExists(policies, managedDiskPolicyName)
		assert.True(t, managedDiskPolicyExists, "Expected to find a managed disk backup policy called %s in vault %s", managedDiskPolicyName, fullVaultName)

		blobStoragePolicyName := fmt.Sprintf("bkpol-%s-blobstorage", vaultName)
		blobStoragePolicyExists := BackupPolicyExists(policies, blobStoragePolicyName)
		assert.True(t, blobStoragePolicyExists, "Expected to find a blob storage backup policy called %s in vault %s", blobStoragePolicyName, fullVaultName)
	}
}
