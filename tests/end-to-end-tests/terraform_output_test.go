package e2e_tests

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

/*
 * TestTerraformOutput tests the output variables of the Terraform deployment.
 */
func TestTerraformOutput(t *testing.T) {
	t.Parallel()

	environment := GetEnvironmentConfiguration(t)

	uniqueId := random.UniqueId()
	resourceGroupName := fmt.Sprintf("rg-nhsbackup-%s", uniqueId)
	resourceGroupLocation := "uksouth"
	backupVaultName := fmt.Sprintf("bvault-nhsbackup-%s", uniqueId)
	backupVaultRedundancy := "LocallyRedundant"

	tags := map[string]string{
		"tagOne":   "tagOneValue",
		"tagTwo":   "tagTwoValue",
		"tagThree": "tagThreeValue",
	}

	// Teardown stage
	// ...

	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, environment.TerraformFolder)

		terraform.Destroy(t, terraformOptions)
	})

	// Setup stage
	// ...

	test_structure.RunTestStage(t, "setup", func() {
		terraformOptions := &terraform.Options{
			TerraformDir: environment.TerraformFolder,

			Vars: map[string]interface{}{
				"resource_group_name":     resourceGroupName,
				"resource_group_location": resourceGroupLocation,
				"backup_vault_name":       backupVaultName,
				"backup_vault_redundancy": backupVaultRedundancy,
				"tags":                    tags,
			},

			BackendConfig: map[string]interface{}{
				"resource_group_name":  environment.TerraformStateResourceGroup,
				"storage_account_name": environment.TerraformStateStorageAccount,
				"container_name":       environment.TerraformStateContainer,
				"key":                  backupVaultName + ".tfstate",
			},
		}

		test_structure.SaveTerraformOptions(t, environment.TerraformFolder, terraformOptions)

		terraform.InitAndApply(t, terraformOptions)
	})

	// Validate stage
	// ...

	test_structure.RunTestStage(t, "validate", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, environment.TerraformFolder)

		actualVaultName := terraform.OutputMap(t, terraformOptions, "backup_vault")["name"]
		assert.Equal(t, backupVaultName, actualVaultName)

		actualVaultLocation := terraform.OutputMap(t, terraformOptions, "backup_vault")["location"]
		assert.Equal(t, resourceGroupLocation, actualVaultLocation)

		actualVaultRedundancy := terraform.OutputMap(t, terraformOptions, "backup_vault")["redundancy"]
		assert.Equal(t, backupVaultRedundancy, actualVaultRedundancy)
	})
}
