package e2e_tests

import (
	"os"
	"testing"

	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
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
