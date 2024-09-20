package e2e_tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

/*
 * TestBasicDeployment tests a basic deployment of the infrastructure using Terraform using the TF output variables.
 */
func TestBasicDeployment(t *testing.T) {
	t.Parallel()

	terraformFolder := test_structure.CopyTerraformFolderToTemp(t, "../../infrastructure", "")
	terraformStateResourceGroup := os.Getenv("TF_STATE_RESOURCE_GROUP")
	terraformStateStorageAccount := os.Getenv("TF_STATE_STORAGE_ACCOUNT")
	terraformStateContainer := os.Getenv("TF_STATE_STORAGE_CONTAINER")

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
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformFolder)

		// Check if the vault name is as expected
		expectedVaultName := fmt.Sprintf("bvault-%s", vaultName)
		actualVaultName := terraform.Output(t, terraformOptions, "vault_name")
		assert.Equal(t, expectedVaultName, actualVaultName)

		// Check if the vault location is as expected
		actualVaultLocation := terraform.Output(t, terraformOptions, "vault_location")
		assert.Equal(t, vaultLocation, actualVaultLocation)

		// Check if the vault redundancy is as expected
		actualVaultRedundancy := terraform.Output(t, terraformOptions, "vault_redundancy")
		assert.Equal(t, vaultRedundancy, actualVaultRedundancy)
	})

	// Teardown stage
	// ...

	test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformFolder)

		terraform.Destroy(t, terraformOptions)
	})
}
