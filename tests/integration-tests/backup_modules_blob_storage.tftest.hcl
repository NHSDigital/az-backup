mock_provider "azurerm" {
  source = "./azurerm"
}

mock_provider "azapi" {
  source = "./azapi"
}

run "setup_tests" {
  module {
    source = "./setup"
  }
}

run "create_blob_storage_backup" {
  command = apply

  module {
    source = "../../infrastructure"
  }

  variables {
    vault_name     = run.setup_tests.vault_name
    vault_location = "uksouth"
    blob_storage_backups = {
      backup1 = {
        backup_name        = "storage1"
        retention_period   = "P7D"
        storage_account_id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.Storage/storageAccounts/sastorage1"
      }
      backup2 = {
        backup_name        = "storage2"
        retention_period   = "P30D"
        storage_account_id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.Storage/storageAccounts/sastorage2"
      }
    }
  }

  assert {
    condition     = length(module.blob_storage_backup) == 2
    error_message = "Number of backup modules not as expected."
  }

  assert {
    condition     = length(module.blob_storage_backup["backup1"].backup_policy.id) > 0
    error_message = "Blob storage backup policy id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup1"].backup_policy.name == "bkpol-${var.vault_name}-blobstorage-storage1"
    error_message = "Blob storage backup policy name not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup1"].backup_policy.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Blob storage backup policy vault id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup1"].backup_policy.operational_default_retention_duration == "P7D"
    error_message = "Blob storage backup policy retention period not as expected."
  }

  assert {
    condition     = length(module.blob_storage_backup["backup1"].backup_instance.id) > 0
    error_message = "Blob storage backup instance id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup1"].backup_instance.name == "bkinst-${var.vault_name}-blobstorage-storage1"
    error_message = "Blob storage backup instance name not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup1"].backup_instance.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Blob storage backup instance vault id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup1"].backup_instance.location == azurerm_data_protection_backup_vault.backup_vault.location
    error_message = "Blob storage backup instance location not as expected."
  }

  assert {
    condition     = length(module.blob_storage_backup["backup1"].backup_instance.storage_account_id) > 0
    error_message = "Blob storage backup instance storage account id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup1"].backup_instance.backup_policy_id == module.blob_storage_backup["backup1"].backup_policy.id
    error_message = "Blob storage backup instance backup policy id not as expected."
  }

  assert {
    condition     = length(module.blob_storage_backup["backup2"].backup_policy.id) > 0
    error_message = "Blob storage backup policy id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup2"].backup_policy.name == "bkpol-${var.vault_name}-blobstorage-storage2"
    error_message = "Blob storage backup policy name not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup2"].backup_policy.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Blob storage backup policy vault id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup2"].backup_policy.operational_default_retention_duration == "P30D"
    error_message = "Blob storage backup policy retention period not as expected."
  }

  assert {
    condition     = length(module.blob_storage_backup["backup2"].backup_instance.id) > 0
    error_message = "Blob storage backup instance id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup2"].backup_instance.name == "bkinst-${var.vault_name}-blobstorage-storage2"
    error_message = "Blob storage backup instance name not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup2"].backup_instance.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Blob storage backup instance vault id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup2"].backup_instance.location == azurerm_data_protection_backup_vault.backup_vault.location
    error_message = "Blob storage backup instance location not as expected."
  }

  assert {
    condition     = length(module.blob_storage_backup["backup2"].backup_instance.storage_account_id) > 0
    error_message = "Blob storage backup instance storage account id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup2"].backup_instance.backup_policy_id == module.blob_storage_backup["backup2"].backup_policy.id
    error_message = "Blob storage backup instance backup policy id not as expected."
  }
}