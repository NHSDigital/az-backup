locals {
    policy_name = "test-backup-policy"
}

run "create_backup_policy" {
  command = plan
  
  module {
      source = "../../../infrastructure/modules/backup_policy/blob_storage"
  }

  variables {
    policy_name = "test-backup-policy"
  }

  # Check that the name is as expected
  assert {
    condition     = azurerm_data_protection_backup_policy_blob_storage.backup_policy.name == "test-backup-policy"
    error_message = "Backup policy name not as expected."
  }
}