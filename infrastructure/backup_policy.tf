module "blob_storage_policy" {
  source           = "./modules/backup_policy/blob_storage"
  policy_name      = "bkpol-${var.vault_name}-blobstorage"
  vault_id         = azurerm_data_protection_backup_vault.backup_vault.id
  retention_period = "P7D" # 7 days
  # NOTE - this blob policy has been configured for operational backup 
  # only, which continuously backs up data and does not need a schedule
}

module "managed_disk_policy" {
  source           = "./modules/backup_policy/managed_disk"
  policy_name      = "bkpol-${var.vault_name}-manageddisk"
  vault_id         = azurerm_data_protection_backup_vault.backup_vault.id
  retention_period = "P7D"                               # 7 days
  backup_intervals = ["R/2024-01-01T00:00:00+00:00/P1D"] # Once per day at 00:00
}
