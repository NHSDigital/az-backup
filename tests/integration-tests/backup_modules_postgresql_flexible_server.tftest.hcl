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

run "create_postgresql_flexible_server_backup" {
  command = apply

  module {
    source = "../../infrastructure"
  }

  variables {
    resource_group_name     = run.setup_tests.resource_group_name
    resource_group_location = "uksouth"
    backup_vault_name       = run.setup_tests.backup_vault_name
    tags                    = run.setup_tests.tags
    postgresql_flexible_server_backups = {
      backup1 = {
        backup_name              = "server1"
        retention_period         = "P7D"
        backup_intervals         = ["R/2024-01-01T00:00:00+00:00/P1D"]
        server_id                = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.DBforPostgreSQL/flexibleServers/server-1"
        server_resource_group_id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group1"
      }
      backup2 = {
        backup_name              = "server2"
        retention_period         = "P30D"
        backup_intervals         = ["R/2024-01-01T00:00:00+00:00/P2D"]
        server_id                = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.DBforPostgreSQL/flexibleServers/server-2"
        server_resource_group_id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group2"
      }
    }
  }

  assert {
    condition     = length(module.postgresql_flexible_server_backup) == 2
    error_message = "Number of backup modules not as expected."
  }

  assert {
    condition     = length(module.postgresql_flexible_server_backup["backup1"].backup_policy.id) > 0
    error_message = "Postgresql flexible server backup policy id not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup1"].backup_policy.name == "bkpol-pgflex-server1"
    error_message = "Postgresql flexible server backup policy name not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup1"].backup_policy.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Postgresql flexible server backup policy vault id not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup1"].backup_policy.default_retention_rule[0].life_cycle[0].duration == "P7D"
    error_message = "Postgresql flexible server backup policy retention period not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup1"].backup_policy.backup_repeating_time_intervals[0] == "R/2024-01-01T00:00:00+00:00/P1D"
    error_message = "Postgresql flexible server backup policy backup intervals not as expected."
  }

  assert {
    condition     = length(module.postgresql_flexible_server_backup["backup1"].backup_instance.id) > 0
    error_message = "Postgresql flexible server backup instance id not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup1"].backup_instance.name == "bkinst-pgflex-server1"
    error_message = "Postgresql flexible server backup instance name not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup1"].backup_instance.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Postgresql flexible server backup instance vault id not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup1"].backup_instance.location == azurerm_data_protection_backup_vault.backup_vault.location
    error_message = "Postgresql flexible server backup instance location not as expected."
  }

  assert {
    condition     = length(module.postgresql_flexible_server_backup["backup1"].backup_instance.server_id) > 0
    error_message = "Postgresql flexible server backup instance server id not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup1"].backup_instance.backup_policy_id == module.postgresql_flexible_server_backup["backup1"].backup_policy.id
    error_message = "Postgresql flexible server backup instance backup policy id not as expected."
  }

  assert {
    condition     = length(module.postgresql_flexible_server_backup["backup2"].backup_policy.id) > 0
    error_message = "Postgresql flexible server backup policy id not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup2"].backup_policy.name == "bkpol-pgflex-server2"
    error_message = "Postgresql flexible server backup policy name not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup2"].backup_policy.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Postgresql flexible server backup policy vault id not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup2"].backup_policy.default_retention_rule[0].life_cycle[0].duration == "P30D"
    error_message = "Postgresql flexible server backup policy retention period not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup2"].backup_policy.backup_repeating_time_intervals[0] == "R/2024-01-01T00:00:00+00:00/P2D"
    error_message = "Postgresql flexible server backup policy backup intervals not as expected."
  }

  assert {
    condition     = length(module.postgresql_flexible_server_backup["backup2"].backup_instance.id) > 0
    error_message = "Postgresql flexible server backup instance id not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup2"].backup_instance.name == "bkinst-pgflex-server2"
    error_message = "Postgresql flexible server backup instance name not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup2"].backup_instance.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Postgresql flexible server backup instance vault id not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup2"].backup_instance.location == azurerm_data_protection_backup_vault.backup_vault.location
    error_message = "Postgresql flexible server backup instance location not as expected."
  }

  assert {
    condition     = length(module.postgresql_flexible_server_backup["backup2"].backup_instance.server_id) > 0
    error_message = "Postgresql flexible server backup instance server id not as expected."
  }

  assert {
    condition     = module.postgresql_flexible_server_backup["backup2"].backup_instance.backup_policy_id == module.postgresql_flexible_server_backup["backup2"].backup_policy.id
    error_message = "Postgresql flexible server backup instance backup policy id not as expected."
  }
}