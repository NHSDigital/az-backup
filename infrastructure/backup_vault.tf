resource "azurerm_data_protection_backup_vault" "backup_vault" {
  name                = "bvault-${var.vault_name}"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = var.vault_location
  datastore_type      = "VaultStore"
  redundancy          = var.vault_redundancy
  soft_delete         = "Off"
  identity {
    type = "SystemAssigned"
  }
}


resource "azapi_update_resource" "immutabilitysettings" {
  type = "Microsoft.DataProtection/backupVaults@2022-11-01-preview"
  resource_id  = azurerm_data_protection_backup_vault.backup_vault.id
   
  
  body = jsonencode({
    properties = {
 
      securitySettings = {
        immutabilitySettings = {
          state = var.vault_immutabilitySettings
        }
      }
    }
    
  })
}