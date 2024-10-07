resource "azurerm_data_protection_backup_vault" "backup_vault" {
  name                = "bvault-${var.vault_name}"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = var.vault_location
  datastore_type      = var.datastore_type
  redundancy          = var.vault_redundancy
  soft_delete         = var.softDeleteSettingsState
  tags                = var.tags
  identity {
    type = "SystemAssigned"
  }
}

resource "azapi_update_resource" "immutabilitysettings" {
  type = "Microsoft.DataProtection/backupVaults@2022-11-01-preview"
  resource_id  = azurerm_data_protection_backup_vault.backup_vault.id
   
  
  body = jsonencode({
    properties = {
      monitoringSettings = {
        azureMonitorAlertSettings = {
          alertsForAllJobFailures = var.monitoringSettings
        }
      }
      securitySettings = {
        immutabilitySettings = {
          state = var.immutabilitySettings
        }
        softDeleteSettings = {
          retentionDurationInDays = var.softDeleteSettingsretentionDays
          state = var.softDeleteSettingsState
        }
      }
      storageSettings = [
        {
          datastoreType = var.datastore_type
          type = var.vault_redundancy
        }
      ]
    }
    
  })
}