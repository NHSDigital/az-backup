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
    vault_name     = run.setup_tests.vault_name
    vault_location = "uksouth"
    managed_disk_backups = {
      backup1 = {
        backup_name      = "disk1"
        retention_period = "P7D"
        backup_intervals = ["R/2024-01-01T00:00:00+00:00/P1D"]
        managed_disk_id  = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.Compute/disks/disk-1"
        managed_disk_resource_group = {
          id   = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group1"
          name = "example-resource-group1"
        }
      }
      backup2 = {
        backup_name      = "disk2"
        retention_period = "P30D"
        backup_intervals = ["R/2024-01-01T00:00:00+00:00/P2D"]
        managed_disk_id  = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.Compute/disks/disk-2"
        managed_disk_resource_group = {
          id   = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group2"
          name = "example-resource-group2"
        }
      }
    }
  }

  assert {
    condition     = length(module.managed_disk_backup) == 2
    error_message = "Number of backup modules not as expected."
  }

  assert {
    condition     = length(module.managed_disk_backup["backup1"].backup_policy.id) > 0
    error_message = "Managed disk backup policy id not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup1"].backup_policy.name == "bkpol-${var.vault_name}-manageddisk-disk1"
    error_message = "Managed disk backup policy name not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup1"].backup_policy.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Managed disk backup policy vault id not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup1"].backup_policy.default_retention_duration == "P7D"
    error_message = "Managed disk backup policy retention period not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup1"].backup_policy.backup_repeating_time_intervals[0] == "R/2024-01-01T00:00:00+00:00/P1D"
    error_message = "Managed disk backup policy backup intervals not as expected."
  }

  assert {
    condition     = length(module.managed_disk_backup["backup1"].backup_instance.id) > 0
    error_message = "Managed disk backup instance id not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup1"].backup_instance.name == "bkinst-${var.vault_name}-manageddisk-disk1"
    error_message = "Managed disk backup instance name not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup1"].backup_instance.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Managed disk backup instance vault id not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup1"].backup_instance.location == azurerm_data_protection_backup_vault.backup_vault.location
    error_message = "Managed disk backup instance location not as expected."
  }

  assert {
    condition     = length(module.managed_disk_backup["backup1"].backup_instance.disk_id) > 0
    error_message = "Managed disk backup instance managed disk id not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup1"].backup_instance.snapshot_resource_group_name == "example-resource-group1"
    error_message = "Managed disk backup instance snapshot resource group not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup1"].backup_instance.backup_policy_id == module.managed_disk_backup["backup1"].backup_policy.id
    error_message = "Managed disk backup instance backup policy id not as expected."
  }

  assert {
    condition     = length(module.managed_disk_backup["backup2"].backup_policy.id) > 0
    error_message = "Managed disk backup policy id not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup2"].backup_policy.name == "bkpol-${var.vault_name}-manageddisk-disk2"
    error_message = "Managed disk backup policy name not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup2"].backup_policy.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Managed disk backup policy vault id not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup2"].backup_policy.default_retention_duration == "P30D"
    error_message = "Managed disk backup policy retention period not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup2"].backup_policy.backup_repeating_time_intervals[0] == "R/2024-01-01T00:00:00+00:00/P2D"
    error_message = "Managed disk backup policy backup intervals not as expected."
  }

  assert {
    condition     = length(module.managed_disk_backup["backup2"].backup_instance.id) > 0
    error_message = "Managed disk backup instance id not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup2"].backup_instance.name == "bkinst-${var.vault_name}-manageddisk-disk2"
    error_message = "Managed disk backup instance name not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup2"].backup_instance.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Managed disk backup instance vault id not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup2"].backup_instance.location == azurerm_data_protection_backup_vault.backup_vault.location
    error_message = "Managed disk backup instance location not as expected."
  }

  assert {
    condition     = length(module.managed_disk_backup["backup2"].backup_instance.disk_id) > 0
    error_message = "Managed disk backup instance managed disk id not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup2"].backup_instance.snapshot_resource_group_name == "example-resource-group2"
    error_message = "Managed disk backup instance snapshot resource group not as expected."
  }

  assert {
    condition     = module.managed_disk_backup["backup2"].backup_instance.backup_policy_id == module.managed_disk_backup["backup2"].backup_policy.id
    error_message = "Managed disk backup instance backup policy id not as expected."
  }
}