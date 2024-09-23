module "blob_storage_backup" {
  for_each           = var.blob_storage_backups
  source             = "./modules/backup/blob_storage"
  vault_id           = azurerm_data_protection_backup_vault.backup_vault.id
  vault_name         = var.vault_name
  vault_location     = var.vault_location
  backup_name        = each.value.backup_name
  retention_period   = each.value.retention_period
  storage_account_id = each.value.storage_account_id
  vault_principal_id = azurerm_data_protection_backup_vault.backup_vault.identity[0].principal_id
}

module "managed_disk_backup" {
  for_each                    = var.managed_disk_backups
  source                      = "./modules/backup/managed_disk"
  vault_id                    = azurerm_data_protection_backup_vault.backup_vault.id
  vault_name                  = var.vault_name
  vault_location              = var.vault_location
  backup_name                 = each.value.backup_name
  retention_period            = each.value.retention_period
  backup_intervals            = each.value.backup_intervals
  managed_disk_id             = each.value.managed_disk_id
  managed_disk_resource_group = each.value.managed_disk_resource_group
  vault_principal_id          = azurerm_data_protection_backup_vault.backup_vault.identity[0].principal_id
}
