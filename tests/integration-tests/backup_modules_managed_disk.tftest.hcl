mock_provider "azurerm" {
  source = "./azurerm"
}

run "setup_tests" {
  module {
    source = "./setup"
  }
}

run "create_managed_disk_backup" {
  command = apply

  module {
    source = "../../infrastructure"
  }

  variables {
    vault_name = run.setup_tests.vault_name
  }

  assert {
    condition     = length(module.managed_disk_backup.id) > 0
    error_message = "Managed disk policy id not as expected."
  }

  assert {
    condition     = module.managed_disk_backup.name == "bkpol-${var.vault_name}-manageddisk"
    error_message = "Managed disk policy name not as expected."
  }

  assert {
    condition     = module.managed_disk_backup.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Managed disk policy vault id not as expected."
  }

  assert {
    condition     = module.managed_disk_backup.retention_period == "P7D"
    error_message = "Managed disk policy retention period not as expected."
  }

  assert {
    condition     = can(module.managed_disk_backup.backup_intervals) && length(module.managed_disk_backup.backup_intervals) == 1 && module.managed_disk_backup.backup_intervals[0] == "R/2024-01-01T00:00:00+00:00/P1D"
    error_message = "Managed disk policy backup intervals not as expected."
  }
}