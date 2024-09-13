mock_provider "azurerm" {
  source = "./azurerm"
}

run "setup_tests" {
  module {
    source = "./setup"
  }
}

run "create_blob_storage_policy" {
  command = apply

  module {
    source = "../infrastructure"
  }

  variables {
    vault_name = run.setup_tests.vault_name
  }

  # Check that the id is as expected
  assert {
    condition     = length(module.blob_storage_policy.id) > 0
    error_message = "Blob storage policy id not as expected."
  }

  # Check that the name is as expected
  assert {
    condition     = module.blob_storage_policy.name == "bkpol-${var.vault_name}-blobstorage"
    error_message = "Blob storage policy name not as expected."
  }

  # Check that the vault id is as expected
  assert {
    condition     = module.blob_storage_policy.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Blob storage policy vault id not as expected."
  }

  # Check that the retention period is as expected
  assert {
    condition     = module.blob_storage_policy.retention_period == "P7D"
    error_message = "Blob storage policy retention period not as expected."
  }
}

run "create_managed_disk_policy" {
  command = apply

  module {
    source = "../infrastructure"
  }

  variables {
    vault_name = run.setup_tests.vault_name
  }

  # Check that the id is as expected
  assert {
    condition     = length(module.managed_disk_policy.id) > 0
    error_message = "Managed disk policy id not as expected."
  }

  # Check that the name is as expected
  assert {
    condition     = module.managed_disk_policy.name == "bkpol-${var.vault_name}-manageddisk"
    error_message = "Managed disk policy name not as expected."
  }

  # Check that the vault id is as expected
  assert {
    condition     = module.managed_disk_policy.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Managed disk policy vault id not as expected."
  }

  # Check that the retention period is as expected
  assert {
    condition     = module.managed_disk_policy.retention_period == "P7D"
    error_message = "Managed disk policy retention period not as expected."
  }

  # Check that the backup intervals is as expected
  assert {
    condition     = can(module.managed_disk_policy.backup_intervals) && length(module.managed_disk_policy.backup_intervals) == 1 && module.managed_disk_policy.backup_intervals[0] == "R/2024-01-01T00:00:00+00:00/P1D"
    error_message = "Managed disk policy backup intervals not as expected."
  }
}