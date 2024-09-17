package e2e_tests

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
)

// Checks if a backup policy exists in the list for the given name
func BackupPolicyExists(list []*armdataprotection.BaseBackupPolicyResource, name string) bool {
	for _, item := range list {
		if *item.Name == name {
			return true
		}
	}

	return false
}
