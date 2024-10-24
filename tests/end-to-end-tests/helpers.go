package e2e_tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
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
 * GetEnvironmentConfiguration gets the environment config that is required to execute a test.
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
 * Gets a resource group for the provided name.
 */
func GetResourceGroup(t *testing.T, subscriptionID string,
	credential *azidentity.ClientSecretCredential, name string) armresources.ResourceGroup {
	// Create a new resource groups client
	client, err := armresources.NewResourceGroupsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create resource group client: %v", err)

	// Get the resource group
	resp, err := client.Get(context.Background(), name, nil)
	assert.NoError(t, err, "Failed to get resource group: %v", err)

	return resp.ResourceGroup
}

/*
 * Gets a role definition for the provided role name.
 */
func GetRoleDefinition(t *testing.T, credential *azidentity.ClientSecretCredential, roleName string) *armauthorization.RoleDefinition {
	roleDefinitionsClient, err := armauthorization.NewRoleDefinitionsClient(credential, nil)
	assert.NoError(t, err, "Failed to create role definition client: %v", err)

	// Create a pager to list role definitions
	filter := fmt.Sprintf("roleName eq '%s'", roleName)
	pager := roleDefinitionsClient.NewListPager("", &armauthorization.RoleDefinitionsClientListOptions{Filter: &filter})

	for pager.More() {
		page, err := pager.NextPage(context.Background())
		assert.NoError(t, err, "Failed to list role definitions")

		for _, roleDefinition := range page.RoleDefinitionListResult.Value {
			if *roleDefinition.Properties.RoleName == roleName {
				return roleDefinition
			}
		}
	}

	return nil
}

/*
 * Gets a role assignment in the provided scope for the provided role definition,
 * that's been assigned to the provided principal id.
 */
func GetRoleAssignment(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string,
	principalId string, roleDefinition *armauthorization.RoleDefinition, scope string) *armauthorization.RoleAssignment {
	roleAssignmentsClient, err := armauthorization.NewRoleAssignmentsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create role assignments client: %v", err)

	// List role assignments for the given scope
	filter := fmt.Sprintf("principalId eq '%s'", principalId)
	pager := roleAssignmentsClient.NewListForScopePager(scope, &armauthorization.RoleAssignmentsClientListForScopeOptions{Filter: &filter})

	// Find the role assignment for the given definition
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		assert.NoError(t, err, "Failed to list role assignments")

		// Check if the role definition is among the assigned roles
		for _, roleAssignment := range page.RoleAssignmentListResult.Value {
			// Use string.contains, as the role definition ID on a role assignment
			// is a longer URI which includes the subscription scope
			if strings.Contains(*roleAssignment.Properties.RoleDefinitionID, *roleDefinition.ID) {
				return roleAssignment
			}
		}
	}

	return nil
}

func GetDiagnosticSettings(t *testing.T, credential *azidentity.ClientSecretCredential, resourceID string, resourceName string) *armmonitor.DiagnosticSettingsResource {
	client, err := armmonitor.NewDiagnosticSettingsClient(credential, nil)
	assert.NoError(t, err, "Failed to create diagnostic settings client: %v", err)

	// List the diagnostic settings for the given resource
	pager := client.NewListPager(resourceID, nil)

	for pager.More() {
		page, err := pager.NextPage(context.Background())
		assert.NoError(t, err, "Failed to list diagnostic settings")

		// We currently only handle when there's only one diagnostic setting per resource
		// ...

		if len(page.Value) == 0 {
			assert.Fail(t, "No diagnostic settings found for resource: %s", resourceName)
		} else if len(page.Value) > 1 {
			assert.Fail(t, "Multiple diagnostic settings found for resource: %s", resourceName)
		} else {
			return page.Value[0]
		}
	}

	return nil
}

/*
 * Gets a backup vault for the provided name.
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
 * Gets a backup instance from the provided list for the provided name
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
func CreateResourceGroup(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string,
	resourceGroupName string, resourceGroupLocation string) armresources.ResourceGroup {
	client, err := armresources.NewResourceGroupsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create resource group client: %v", err)

	log.Printf("Creating resource group %s in location %s", resourceGroupName, resourceGroupLocation)

	resp, err := client.CreateOrUpdate(
		context.Background(),
		resourceGroupName,
		armresources.ResourceGroup{
			Location: &resourceGroupLocation,
		},
		nil,
	)
	assert.NoError(t, err, "Failed to create resource group: %v", err)

	log.Printf("Resource group %s created successfully", resourceGroupName)

	return resp.ResourceGroup
}

/*
 * Creates a Log Analytics workspace that can be used for testing purposes.
 */
func CreateLogAnalyticsWorkspace(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string,
	resourceGroupName string, workspaceName string, workspaceLocation string) armoperationalinsights.Workspace {
	client, err := armoperationalinsights.NewWorkspacesClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create Log Analytics workspace client: %v", err)

	log.Printf("Creating log analytics workspace %s in location %s", workspaceName, workspaceLocation)

	pollerResp, err := client.BeginCreateOrUpdate(
		context.Background(),
		resourceGroupName,
		workspaceName,
		armoperationalinsights.Workspace{
			Location: &workspaceLocation,
		},
		nil,
	)
	assert.NoError(t, err, "Failed to begin creating log analytics workspace: %v", err)

	// Wait for the creation to complete
	resp, err := pollerResp.PollUntilDone(context.Background(), nil)
	assert.NoError(t, err, "Failed to create log analytics workspace: %v", err)

	log.Printf("Log analytics workspace %s created successfully", workspaceName)

	return resp.Workspace
}

/*
 * Creates a storage account that can be used for testing purposes.
 */
func CreateStorageAccount(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string,
	resourceGroupName string, storageAccountName string, storageAccountLocation string) armstorage.Account {
	client, err := armstorage.NewAccountsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create storage account client: %v", err)

	log.Printf("Creating storage account %s in location %s", storageAccountName, storageAccountLocation)

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

	log.Printf("Storage account %s created successfully", storageAccountName)

	return resp.Account
}

/*
 * Creates a storage account container that can be used for testing purposes.
 */
func CreateStorageAccountContainer(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string,
	resourceGroupName string, storageAccountName string, containerName string) armstorage.BlobContainer {
	containerClient, err := armstorage.NewBlobContainersClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create container client: %v", err)

	resp, err := containerClient.Create(
		context.Background(),
		resourceGroupName,
		storageAccountName,
		containerName,
		armstorage.BlobContainer{},
		nil,
	)
	assert.NoError(t, err, "Failed to create container: %v", err)

	log.Printf("Container '%s' created successfully in storage account %s", containerName, storageAccountName)

	return resp.BlobContainer
}

/*
 * Creates a managed disk that can be used for testing purposes.
 */
func CreateManagedDisk(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string,
	resourceGroupName string, diskName string, diskLocation string, diskSizeGB int32) armcompute.Disk {
	client, err := armcompute.NewDisksClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create disks client: %v", err)

	log.Printf("Creating managed disk %s in location %s", diskName, diskLocation)

	pollerResp, err := client.BeginCreateOrUpdate(
		context.Background(),
		resourceGroupName,
		diskName,
		armcompute.Disk{
			Location: &diskLocation,
			SKU: &armcompute.DiskSKU{
				Name: to.Ptr(armcompute.DiskStorageAccountTypesStandardLRS),
			},
			Properties: &armcompute.DiskProperties{
				DiskSizeGB:   &diskSizeGB,
				CreationData: &armcompute.CreationData{CreateOption: to.Ptr(armcompute.DiskCreateOptionEmpty)},
			},
		},
		nil,
	)
	assert.NoError(t, err, "Failed to begin creating managed disk: %v", err)

	// Wait for the creation to complete
	resp, err := pollerResp.PollUntilDone(context.Background(), nil)
	assert.NoError(t, err, "Failed to create managed disk: %v", err)

	log.Printf("Managed disk %s created successfully", diskName)

	return resp.Disk
}

/*
 * Creates a postgresql flexible server that can be used for testing purposes.
 */
func CreatePostgresqlFlexibleServer(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string,
	resourceGroupName string, serverName string, serverLocation string, storageSizeGB int32) armpostgresqlflexibleservers.Server {
	client, err := armpostgresqlflexibleservers.NewServersClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create servers client: %v", err)

	log.Printf("Creating postgresql flexible server %s in location %s", serverName, serverLocation)

	pollerResp, err := client.BeginCreate(
		context.Background(),
		resourceGroupName,
		serverName,
		armpostgresqlflexibleservers.Server{
			Location: &serverLocation,
			SKU: &armpostgresqlflexibleservers.SKU{
				Name: to.Ptr("Standard_B1ms"),
				Tier: to.Ptr(armpostgresqlflexibleservers.SKUTierBurstable),
			},
			Properties: &armpostgresqlflexibleservers.ServerProperties{
				AdministratorLogin:         to.Ptr("supersecurelogin"),
				AdministratorLoginPassword: to.Ptr("supersecurepassword"),
				Version:                    to.Ptr(armpostgresqlflexibleservers.ServerVersionFourteen),
				Storage: &armpostgresqlflexibleservers.Storage{
					StorageSizeGB: &storageSizeGB,
				},
			},
		},
		nil,
	)
	assert.NoError(t, err, "Failed to begin creating postgresql flexible server: %v", err)

	// Wait for the creation to complete
	resp, err := pollerResp.PollUntilDone(context.Background(), nil)
	assert.NoError(t, err, "Failed to create postgresql flexible server: %v", err)

	log.Printf("Postgresql flexible server %s created successfully", serverName)

	return resp.Server
}

/*
 * Deletes a resource group.
 */
func DeleteResourceGroup(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string) {
	client, err := armresources.NewResourceGroupsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create resource group client: %v", err)

	log.Printf("Deleting resource group %s", resourceGroupName)

	pollerResp, err := client.BeginDelete(context.Background(), resourceGroupName, nil)
	assert.NoError(t, err, "Failed to delete resource group: %v", err)

	// Wait for the creation to complete
	_, err = pollerResp.PollUntilDone(context.Background(), nil)
	assert.NoError(t, err, "Failed to create storage account: %v", err)

	log.Printf("Resource group %s deleted successfully", resourceGroupName)
}

/*
 * Deletes the backup instance for the provided backup vault and instance name.
 */
func DeleteBackupInstance(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, backupVaultName string, backupInstanceName string) error {
	client, err := armdataprotection.NewBackupInstancesClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create data protection client: %v", err)

	poller, err := client.BeginDelete(context.Background(), resourceGroupName, backupVaultName, backupInstanceName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete backup instance: %w", err)
	}

	_, err = poller.PollUntilDone(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to delete backup instance: %w", err)
	}

	log.Printf("Backup instance '%s' deleted successfully", backupInstanceName)
	return nil
}

/*
 * Creates a test file that can be used for test purposes.
 */
func CreateTestFile(t *testing.T) *os.File {
	testFile, err := os.CreateTemp("", "test.txt")
	assert.NoError(t, err, "Failed to test file: %v", err)
	defer os.Remove(testFile.Name())

	content := []byte("This is a test file for upload.")
	testFile.Write(content)
	testFile.Close()
	return testFile
}

/*
 * Uploads a file to blob storage account
 */
func UploadFileToStorageAccount(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, storageAccountName string, containerName string, filePath string) {
	serviceClient, err := azblob.NewClient(fmt.Sprintf("https://%s.blob.core.windows.net/", storageAccountName), credential, nil)
	assert.NoError(t, err, "Failed to create service client: %v", err)

	file, err := os.Open(filePath)
	assert.NoError(t, err, "Failed to open file: %v", err)
	defer file.Close()

	_, err = serviceClient.UploadFile(context.Background(), containerName, filepath.Base(file.Name()), file, nil)
	assert.NoError(t, err, "Failed to upload file: %v", err)

	log.Printf("File '%s' uploaded successfully to container '%s' in storage account '%s'", filePath, containerName, storageAccountName)
}

/*
 * Updates the immutability setting on a backup vault.
 */
func UpdateBackupVaultImmutability(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, backupVaultName string, immutabilitySettings armdataprotection.ImmutabilitySettings) {
	client, err := armdataprotection.NewBackupVaultsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create data protection client: %v", err)

	// Set the immutability setting on the backup vault
	_, err = client.BeginUpdate(context.Background(), resourceGroupName, backupVaultName, armdataprotection.PatchResourceRequestInput{
		Properties: &armdataprotection.PatchBackupVaultInput{
			SecuritySettings: &armdataprotection.SecuritySettings{
				ImmutabilitySettings: &immutabilitySettings,
			},
		},
	}, nil)
	assert.NoError(t, err, "Failed to set immutability setting on backup vault: %v", err)

	log.Printf("Immutability setting updated on backup vault '%s'", backupVaultName)
}

/*
 * Begins an ad-hoc backup for the provided backup instance name.
 */
func BeginAdHocBackup(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, backupVaultName string, backupInstanceName string) {
	instancesClient, err := armdataprotection.NewBackupInstancesClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create backup instances client: %v", err)

	poller, err := instancesClient.BeginAdhocBackup(context.Background(), resourceGroupName, backupVaultName, backupInstanceName, armdataprotection.TriggerBackupRequest{
		BackupRuleOptions: &armdataprotection.AdHocBackupRuleOptions{
			RuleName:      to.Ptr("BackupIntervals"),
			TriggerOption: &armdataprotection.AdhocBackupTriggerOption{},
		},
	}, nil)
	assert.NoError(t, err, "Failed to begin ad-hoc backup: %v", err)

	resp, err := poller.PollUntilDone(context.Background(), nil)
	assert.NoError(t, err, "Failed to poll ad-hoc backup status: %v", err)
	assert.NotNil(t, *resp.JobID, "Expected a job ID to be returned")

	jobClient, err := armdataprotection.NewJobsClient(subscriptionID, credential, nil)
	assert.NoError(t, err, "Failed to create backup jobs client: %v", err)

	jobId := strings.Split(*resp.JobID, "/")[len(strings.Split(*resp.JobID, "/"))-1]

	for {
		jobResp, err := jobClient.Get(context.Background(), resourceGroupName, backupVaultName, jobId, nil)
		assert.NoError(t, err, "Failed to get backup job status: %v", err)

		if *jobResp.Properties.Status != "InProgress" {
			assert.Equal(t, "Completed", *jobResp.Properties.Status, "Backup job did not succeed")
			break
		}

		log.Printf("Backup job '%s' is still in progress...", jobId)

		time.Sleep(10 * time.Second)
	}

	log.Printf("Ad-hoc backup '%s' completed successfully", backupInstanceName)
}
