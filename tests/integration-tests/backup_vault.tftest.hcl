mock_provider "azurerm" {
  source = "./azurerm"
}

run "setup_tests" {
  module {
    source = "./setup"
  }
}

run "create_backup_vault" {
  command = apply

  module {
    source = "../../infrastructure"
  }

  variables {
    resource_group_name       = run.setup_tests.resource_group_name
    resource_group_location   = "uksouth"
    backup_vault_name         = run.setup_tests.backup_vault_name
    backup_vault_redundancy   = "LocallyRedundant"
    backup_vault_immutability = "Unlocked"
    tags                      = run.setup_tests.tags
  }

  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.name == var.backup_vault_name
    error_message = "Backup vault name not as expected."
  }

  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.resource_group_name == local.resource_group.name
    error_message = "Resource group not as expected."
  }

  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.location == local.resource_group.location
    error_message = "Backup vault location not as expected."
  }

  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.datastore_type == "VaultStore"
    error_message = "Backup vault datastore type not as expected."
  }

  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.redundancy == var.backup_vault_redundancy
    error_message = "Backup vault redundancy not as expected."
  }

  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.soft_delete == "Off"
    error_message = "Backup vault soft delete not as expected."
  }

  assert {
    condition     = length(azurerm_data_protection_backup_vault.backup_vault.identity[0].principal_id) > 0
    error_message = "Backup vault identity not as expected."
  }

  assert {
    condition     = length(azurerm_data_protection_backup_vault.backup_vault.tags) == length(run.setup_tests.tags)
    error_message = "Tags not as expected."
  }

  assert {
    condition = alltrue([
      for tag_key, tag_value in run.setup_tests.tags :
      lookup(azurerm_data_protection_backup_vault.backup_vault.tags, tag_key, null) == tag_value
    ])
    error_message = "Tags not as expected."
  }
  
  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.immutability == var.backup_vault_immutability
    error_message = "Backup vault immutability not as expected."
  }
}

run "configure_vault_diagnostics_when_enabled" {
  command = apply

  module {
    source = "../../infrastructure"
  }

  variables {
    resource_group_name        = run.setup_tests.resource_group_name
    resource_group_location    = "uksouth"
    backup_vault_name          = run.setup_tests.backup_vault_name
    log_analytics_workspace_id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.OperationalInsights/workspaces/workspace1"
    tags                       = run.setup_tests.tags
  }

  assert {
    condition     = length(azurerm_monitor_diagnostic_setting.backup_vault) == 1
    error_message = "Backup vault diagnostic settings not as expected."
  }

  assert {
    condition     = azurerm_monitor_diagnostic_setting.backup_vault[0].target_resource_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Backup vault diagnostic setting target resource id not as expected."
  }

  assert {
    condition     = length(azurerm_monitor_diagnostic_setting.backup_vault[0].log_analytics_workspace_id) > 0
    error_message = "Backup vault diagnostic setting log analytics workspace id not as expected."
  }

  assert {
    condition     = length(azurerm_monitor_diagnostic_setting.backup_vault[0].enabled_log) == length(local.backup_vault_diagnostics_log_categories)
    error_message = "Backup vault diagnostic setting enabled logs not as expected."
  }

  assert {
    condition     = alltrue([for enabled_log in azurerm_monitor_diagnostic_setting.backup_vault[0].enabled_log : contains(local.backup_vault_diagnostics_log_categories, enabled_log.category)])
    error_message = "Backup vault diagnostic setting enabled logs not as expected."
  }

  assert {
    condition     = length(azurerm_monitor_diagnostic_setting.backup_vault[0].metric) == length(local.backup_vault_diagnostics_metric_categories)
    error_message = "Backup vault diagnostic setting metrics not as expected."
  }

  assert {
    condition     = alltrue([for metric in azurerm_monitor_diagnostic_setting.backup_vault[0].metric : contains(local.backup_vault_diagnostics_metric_categories, metric.category)])
    error_message = "Backup vault diagnostic setting metrics not as expected."
  }
}

run "configure_vault_diagnostics_when_disabled" {
  command = apply

  module {
    source = "../../infrastructure"
  }

  variables {
    resource_group_name     = run.setup_tests.resource_group_name
    resource_group_location = "uksouth"
    backup_vault_name       = run.setup_tests.backup_vault_name
    tags                    = run.setup_tests.tags
  }

  assert {
    condition     = length(azurerm_monitor_diagnostic_setting.backup_vault) == 0
    error_message = "Backup vault diagnostic settings not as expected."
  }
}
