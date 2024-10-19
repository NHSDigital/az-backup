resource "azurerm_data_protection_backup_vault" "backup_vault" {
  name                = var.backup_vault_name
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  datastore_type      = "VaultStore"
  redundancy          = var.backup_vault_redundancy
  soft_delete         = "Off"
  tags                = var.tags
  identity {
    type = "SystemAssigned"
  }
}

resource "azapi_update_resource" "backup_vault_settings" {
  type        = "Microsoft.DataProtection/backupVaults@2022-11-01-preview"
  resource_id = azurerm_data_protection_backup_vault.backup_vault.id
  body = jsonencode({
    properties = {
      securitySettings = {
        immutabilitySettings = {
          state = var.backup_vault_immutability
        }
      }
    }
  })
}

locals {
  backup_vault_diagnostics_log_categories = toset([
    "AddonAzureBackupJobs",
    "AddonAzureBackupPolicy",
    "AddonAzureBackupProtectedInstance",
    "CoreAzureBackup"
  ])

  backup_vault_diagnostics_metric_categories = toset([
    "Health"
  ])
}

resource "azurerm_monitor_diagnostic_setting" "backup_vault" {
  count                      = length(var.log_analytics_workspace_id) > 0 ? 1 : 0
  name                       = "${var.backup_vault_name}-diagnostic-settings"
  target_resource_id         = azurerm_data_protection_backup_vault.backup_vault.id
  log_analytics_workspace_id = var.log_analytics_workspace_id

  dynamic "enabled_log" {
    for_each = toset(local.backup_vault_diagnostics_log_categories)
    content {
      category = enabled_log.key
    }
  }

  dynamic "metric" {
    for_each = toset(local.backup_vault_diagnostics_metric_categories)
    content {
      category = metric.key
    }
  }
}
