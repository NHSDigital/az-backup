module "blob_storage_backup" {
  for_each                   = var.blob_storage_backups
  source                     = "./modules/backup/blob_storage"
  vault                      = azurerm_data_protection_backup_vault.backup_vault
  backup_name                = each.value.backup_name
  retention_period           = each.value.retention_period
  backup_intervals           = each.value.backup_intervals
  storage_account_id         = each.value.storage_account_id
  storage_account_containers = each.value.storage_account_containers
}

module "managed_disk_backup" {
  for_each                          = var.managed_disk_backups
  source                            = "./modules/backup/managed_disk"
  vault                             = azurerm_data_protection_backup_vault.backup_vault
  backup_name                       = each.value.backup_name
  retention_period                  = each.value.retention_period
  backup_intervals                  = each.value.backup_intervals
  managed_disk_id                   = each.value.managed_disk_id
  managed_disk_resource_group       = each.value.managed_disk_resource_group
  assign_resource_group_level_roles = each.key == keys(var.managed_disk_backups)[0] ? true : false
}

module "postgresql_flexible_server_backup" {
  for_each                          = var.postgresql_flexible_server_backups
  source                            = "./modules/backup/postgresql_flexible_server"
  vault                             = azurerm_data_protection_backup_vault.backup_vault
  backup_name                       = each.value.backup_name
  retention_period                  = each.value.retention_period
  backup_intervals                  = each.value.backup_intervals
  server_id                         = each.value.server_id
  server_resource_group_id          = each.value.server_resource_group_id
  assign_resource_group_level_roles = each.key == keys(var.postgresql_flexible_server_backups)[0] ? true : false
}
