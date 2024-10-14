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

	vaultName := random.UniqueId()
	vaultLocation := "uksouth"
	vaultRedundancy := "LocallyRedundant"

	tags := map[string]string{
		"environment":         "production",
		"owner":               "owner_name",
		"created_by":          "creator_name",
		"costing_pcode":       "pcode_value",
		"ch_cost_centre":      "cost_centre_value",
		"project":             "project_name",
		"service_level":       "gold",
		"directorate":         "directorate_name",
		"sub_directorate":     "sub_directorate_name",
		"data_classification": "3",
		"service_product":     "product_name",
		"team":                "team_name",
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
				"vault_name":       vaultName,
				"vault_location":   vaultLocation,
				"vault_redundancy": vaultRedundancy,
				"tags":             tags,
			},

			BackendConfig: map[string]interface{}{
				"resource_group_name":  environment.TerraformStateResourceGroup,
				"storage_account_name": environment.TerraformStateStorageAccount,
				"container_name":       environment.TerraformStateContainer,
				"key":                  vaultName + ".tfstate",
			},
		}

		test_structure.SaveTerraformOptions(t, environment.TerraformFolder, terraformOptions)

		terraform.InitAndApply(t, terraformOptions)
	})

	// Validate stage
	// ...

	test_structure.RunTestStage(t, "validate", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, environment.TerraformFolder)

		expectedVaultName := fmt.Sprintf("bvault-%s", vaultName)
		actualVaultName := terraform.OutputMap(t, terraformOptions, "backup_vault")["name"]
		assert.Equal(t, expectedVaultName, actualVaultName)

		actualVaultLocation := terraform.OutputMap(t, terraformOptions, "backup_vault")["location"]
		assert.Equal(t, vaultLocation, actualVaultLocation)

		actualVaultRedundancy := terraform.OutputMap(t, terraformOptions, "backup_vault")["redundancy"]
		assert.Equal(t, vaultRedundancy, actualVaultRedundancy)
	})
}
