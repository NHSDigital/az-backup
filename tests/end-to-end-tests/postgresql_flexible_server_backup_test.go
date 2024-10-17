package e2e_tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

type TestPostgresqlFlexibleServerBackupExternalResources struct {
	ResourceGroup               armresources.ResourceGroup
	PostgresqlFlexibleServerOne armpostgresqlflexibleservers.Server
	PostgresqlFlexibleServerTwo armpostgresqlflexibleservers.Server
}

/*
 * Creates resources which are "external" to the az-backup module, and models
 * what would be backed up in a real scenario.
 */
func setupExternalResourcesForPostgresqlFlexibleServerBackupTest(t *testing.T, credential *azidentity.ClientSecretCredential, subscriptionID string, resourceGroupName string, resourceGroupLocation string, uniqueId string) *TestPostgresqlFlexibleServerBackupExternalResources {
	externalResourceGroupName := fmt.Sprintf("%s-external", resourceGroupName)
	resourceGroup := CreateResourceGroup(t, credential, subscriptionID, externalResourceGroupName, resourceGroupLocation)

	PostgresqlFlexibleServerOneName := fmt.Sprintf("pgflexserver-%s-external-1", strings.ToLower(uniqueId))
	PostgresqlFlexibleServerOne := CreatePostgresqlFlexibleServer(t, credential, subscriptionID, resourceGroupName, PostgresqlFlexibleServerOneName, resourceGroupLocation, int32(32))

	PostgresqlFlexibleServerTwoName := fmt.Sprintf("pgflexserver-%s-external-2", strings.ToLower(uniqueId))
	PostgresqlFlexibleServerTwo := CreatePostgresqlFlexibleServer(t, credential, subscriptionID, resourceGroupName, PostgresqlFlexibleServerTwoName, resourceGroupLocation, int32(32))

	externalResources := &TestPostgresqlFlexibleServerBackupExternalResources{
		ResourceGroup:               resourceGroup,
		PostgresqlFlexibleServerOne: PostgresqlFlexibleServerOne,
		PostgresqlFlexibleServerTwo: PostgresqlFlexibleServerTwo,
	}

	return externalResources
}

/*
 * TestPostgresqlFlexibleServerBackup tests the deployment of a backup vault and backup policies for postgresql flexible servers.
 */
func TestPostgresqlFlexibleServerBackup(t *testing.T) {
	t.Parallel()

	environment := GetEnvironmentConfiguration(t)
	credential := GetAzureCredential(t, environment)

	uniqueId := random.UniqueId()
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", uniqueId)
	resourceGroupLocation := "uksouth"
	backupVaultName := fmt.Sprintf("bvault-nhsbackup-%s", uniqueId)
	backupVaultRedundancy := "LocallyRedundant"

	externalResources := setupExternalResourcesForPostgresqlFlexibleServerBackupTest(t, credential, environment.SubscriptionID, resourceGroupName, resourceGroupLocation, uniqueId)

	// A map of backups which we'll use to apply the TF module, and then validate the
	// policies have been created correctly
	PostgresqlFlexibleServerBackups := map[string]map[string]interface{}{
		"backup1": {
			"backup_name":              "server1",
			"retention_period":         "P7D",
			"backup_intervals":         []string{"R/2024-01-01T00:00:00+00:00/P1D"},
			"server_id":                *externalResources.PostgresqlFlexibleServerOne.ID,
			"server_resource_group_id": *externalResources.ResourceGroup.ID,
		},
		"backup2": {
			"backup_name":              "server2",
			"retention_period":         "P30D",
			"backup_intervals":         []string{"R/2024-01-01T00:00:00+00:00/P2D"},
			"server_id":                *externalResources.PostgresqlFlexibleServerTwo.ID,
			"server_resource_group_id": *externalResources.ResourceGroup.ID,
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
				"resource_group_name":                resourceGroupName,
				"resource_group_location":            resourceGroupLocation,
				"backup_vault_name":                  backupVaultName,
				"backup_vault_redundancy":            backupVaultRedundancy,
				"postgresql_flexible_server_backups": PostgresqlFlexibleServerBackups,
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

		assert.Equal(t, len(PostgresqlFlexibleServerBackups), len(backupPolicies), "Expected to find %2 backup policies in vault", len(PostgresqlFlexibleServerBackups))
		assert.Equal(t, len(PostgresqlFlexibleServerBackups), len(backupInstances), "Expected to find %2 backup instances in vault", len(PostgresqlFlexibleServerBackups))

		for _, backup := range PostgresqlFlexibleServerBackups {
			backupName := backup["backup_name"].(string)
			retentionPeriod := backup["retention_period"].(string)
			backupIntervals := backup["backup_intervals"].([]string)
			ServerId := backup["server_id"].(string)
			ServerResourceGroupId := backup["server_resource_group_id"].(string)

			// Validate backup policy
			backupPolicyName := fmt.Sprintf("bkpol-%s-pgflexserver-%s", backupVaultName, backupName)
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
			backupInstanceName := fmt.Sprintf("bkinst-%s-pgflexserver-%s", backupVaultName, backupName)
			backupInstance := GetBackupInstanceForName(backupInstances, backupInstanceName)
			assert.NotNil(t, backupInstance, "Expected to find a backup policy called %s", backupInstanceName)
			assert.Equal(t, ServerId, *backupInstance.Properties.DataSourceInfo.ResourceID, "Expected the backup instance source resource ID to be %s", ServerId)
			assert.Equal(t, *backupPolicy.ID, *backupInstance.Properties.PolicyInfo.PolicyID, "Expected the backup instance policy ID to be %s", backupPolicy.ID)

			// Validate role assignments
			readerRoleDefinition := GetRoleDefinition(t, credential, "Reader")
			readerRoleAssignment := GetRoleAssignment(t, credential, environment.SubscriptionID, *backupVault.Identity.PrincipalID, readerRoleDefinition, ServerResourceGroupId)
			assert.NotNil(t, readerRoleAssignment, "Expected to find role assignment %s for principal %s on scope %s", readerRoleDefinition.Name, *backupVault.Identity.PrincipalID, ServerResourceGroupId)

			longTermRetentionBackupRoleDefinition := GetRoleDefinition(t, credential, "PostgreSQL Flexible Server Long Term Retention Backup Role")
			longTermRetentionBackupRoleAssignment := GetRoleAssignment(t, credential, environment.SubscriptionID, *backupVault.Identity.PrincipalID, longTermRetentionBackupRoleDefinition, ServerId)
			assert.NotNil(t, longTermRetentionBackupRoleAssignment, "Expected to find role assignment %s for principal %s on scope %s", longTermRetentionBackupRoleDefinition.Name, *backupVault.Identity.PrincipalID, ServerId)
		}
	})
}
