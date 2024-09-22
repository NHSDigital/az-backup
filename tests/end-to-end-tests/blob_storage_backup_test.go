package e2e_tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

/*
 * TestBlobStorageBackup tests the end to end backup process for a blob storage account, including
 * assigning the Azure policy definition, creating a tag on a backup resource and tetsing the backup
 * instance is created and working.
 */
func TestBlobStorageBackup(t *testing.T) {
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
	storageAccountName := fmt.Sprintf("satest%s", random.UniqueId())

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
		//fullVaultName := fmt.Sprintf("bvault-%s", vaultName)

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

		// Create a storage account with the tag that should trigger the Azure policy
		CreateStorageAccount(t, subscriptionID, cred, resourceGroupName, storageAccountName, vaultLocation)

		// Verify that a backup instance has been created on the vault

		// Verify that a role assignment on the storage account has been granted to the vault identity

		// Upload a blob to the storage account

		// Check that the blob has been backed up
	})

	// Teardown stage
	// ...

	test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformFolder)

		terraform.Destroy(t, terraformOptions)
	})
}

/*
 * Creates a storage account to be used in the test.
 */
func CreateStorageAccount(t *testing.T, subscriptionID string,
	cred *azidentity.ClientSecretCredential, resourceGroupName string, storageAccountName string, vaultLocation string) {
	// Create a new storage account client
	client, err := armstorage.NewAccountsClient(subscriptionID, cred, nil)
	assert.NoError(t, err, "Failed to create storage account client: %v", err)

	// Create the storage account
	pollerResp, err := client.BeginCreate(
		context.Background(),
		resourceGroupName,
		storageAccountName,
		armstorage.AccountCreateParameters{
			SKU: &armstorage.SKU{
				Name: to.Ptr(armstorage.SKUNameStandardLRS),
			},
			Kind:     to.Ptr(armstorage.KindStorageV2),
			Location: &vaultLocation,
		},
		nil,
	)
	assert.NoError(t, err, "Failed to begin creating storage account: %v", err)

	// Wait for the creation to complete
	resp, err := pollerResp.PollUntilDone(context.Background(), nil)
	assert.NoError(t, err, "Failed to create storage account: %v", err)

	fmt.Printf("Storage account %s created successfully\n", *resp.Name)
}
