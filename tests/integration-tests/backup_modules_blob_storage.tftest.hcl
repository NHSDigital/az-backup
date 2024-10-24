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
    resource_group_name     = run.setup_tests.resource_group_name
    resource_group_location = "uksouth"
    backup_vault_name       = run.setup_tests.backup_vault_name
    tags                    = run.setup_tests.tags
    blob_storage_backups = {
      backup1 = {
        backup_name                = "storage1"
        retention_period           = "P1D"
        backup_intervals           = ["R/2024-01-01T00:00:00+00:00/P1D"]
        storage_account_id         = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.Storage/storageAccounts/sastorage1"
        storage_account_containers = ["container1"]
      }
      backup2 = {
        backup_name                = "storage2"
        retention_period           = "P7D"
        backup_intervals           = ["R/2024-01-01T00:00:00+00:00/P2D"]
        storage_account_id         = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.Storage/storageAccounts/sastorage2"
        storage_account_containers = ["container2"]
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
    condition     = module.blob_storage_backup["backup1"].backup_policy.name == "bkpol-blob-storage1"
    error_message = "Blob storage backup policy name not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup1"].backup_policy.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Blob storage backup policy vault id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup1"].backup_policy.vault_default_retention_duration == "P1D"
    error_message = "Blob storage backup policy retention period not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup1"].backup_policy.backup_repeating_time_intervals[0] == "R/2024-01-01T00:00:00+00:00/P1D"
    error_message = "Blob storage backup policy backup intervals not as expected."
  }

  assert {
    condition     = length(module.blob_storage_backup["backup1"].backup_instance.id) > 0
    error_message = "Blob storage backup instance id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup1"].backup_instance.name == "bkinst-blob-storage1"
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
    condition     = module.blob_storage_backup["backup1"].backup_instance.storage_account_container_names[0] == "container1"
    error_message = "Blob storage backup instance storage account containers not as expected."
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
    condition     = module.blob_storage_backup["backup2"].backup_policy.name == "bkpol-blob-storage2"
    error_message = "Blob storage backup policy name not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup2"].backup_policy.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Blob storage backup policy vault id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup2"].backup_policy.vault_default_retention_duration == "P7D"
    error_message = "Blob storage backup policy retention period not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup2"].backup_policy.backup_repeating_time_intervals[0] == "R/2024-01-01T00:00:00+00:00/P2D"
    error_message = "Blob storage backup policy backup intervals not as expected."
  }

  assert {
    condition     = length(module.blob_storage_backup["backup2"].backup_instance.id) > 0
    error_message = "Blob storage backup instance id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup2"].backup_instance.name == "bkinst-blob-storage2"
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
    condition     = module.blob_storage_backup["backup2"].backup_instance.storage_account_container_names[0] == "container2"
    error_message = "Blob storage backup instance storage account containers not as expected."
  }

  assert {
    condition     = module.blob_storage_backup["backup2"].backup_instance.backup_policy_id == module.blob_storage_backup["backup2"].backup_policy.id
    error_message = "Blob storage backup instance backup policy id not as expected."
  }
}

run "validate_retention_period" {
  command = plan

  module {
    source = "../../infrastructure"
  }

  variables {
    resource_group_name     = run.setup_tests.resource_group_name
    resource_group_location = "uksouth"
    backup_vault_name       = run.setup_tests.backup_vault_name
    tags                    = run.setup_tests.tags
    blob_storage_backups = {
      backup1 = {
        backup_name                = "storage1"
        retention_period           = "P30D"
        backup_intervals           = ["R/2024-01-01T00:00:00+00:00/P1D"]
        storage_account_id         = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.Storage/storageAccounts/sastorage1"
        storage_account_containers = ["container1"]
      }
    }
  }

  expect_failures = [
    var.blob_storage_backups,
  ]
}

run "validate_retention_period_with_extended_retention" {
  command = plan

  module {
    source = "../../infrastructure"
  }

  variables {
    resource_group_name     = run.setup_tests.resource_group_name
    resource_group_location = "uksouth"
    backup_vault_name       = run.setup_tests.backup_vault_name
    tags                    = run.setup_tests.tags
    use_extended_retention  = true
    blob_storage_backups = {
      backup1 = {
        backup_name                = "storage1"
        retention_period           = "P30D"
        backup_intervals           = ["R/2024-01-01T00:00:00+00:00/P1D"]
        storage_account_id         = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.Storage/storageAccounts/sastorage1"
        storage_account_containers = ["container1"]
      }
    }
  }

  assert {
    condition     = length(module.blob_storage_backup) == 1
    error_message = "Number of backup modules not as expected."
  }
}

run "validate_backup_intervals" {
  command = plan

  module {
    source = "../../infrastructure"
  }

  variables {
    resource_group_name     = run.setup_tests.resource_group_name
    resource_group_location = "uksouth"
    backup_vault_name       = run.setup_tests.backup_vault_name
    tags                    = run.setup_tests.tags
    blob_storage_backups = {
      backup1 = {
        backup_name                = "storage1"
        retention_period           = "P7D"
        backup_intervals           = []
        storage_account_id         = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.Storage/storageAccounts/sastorage1"
        storage_account_containers = ["container1"]
      }
    }
  }

  expect_failures = [
    var.blob_storage_backups,
  ]
}

run "validate_storage_account_containers" {
  command = plan

  module {
    source = "../../infrastructure"
  }

  variables {
    resource_group_name     = run.setup_tests.resource_group_name
    resource_group_location = "uksouth"
    backup_vault_name       = run.setup_tests.backup_vault_name
    tags                    = run.setup_tests.tags
    blob_storage_backups = {
      backup1 = {
        backup_name                = "storage1"
        retention_period           = "P7D"
        backup_intervals           = ["R/2024-01-01T00:00:00+00:00/P1D"]
        storage_account_id         = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.Storage/storageAccounts/sastorage1"
        storage_account_containers = []
      }
    }
  }

  expect_failures = [
    var.blob_storage_backups,
  ]
}
