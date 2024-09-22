module "blob_storage_backup" {
  source           = "./modules/backup/blob_storage"
  vault_id         = azurerm_data_protection_backup_vault.backup_vault.id
  vault_name       = var.vault_name
  subscription_id  = data.azurerm_subscription.current.id
  backup_name      = "blobstorage"
  retention_period = "P7D" # 7 days
  # NOTE - this blob policy has been configured for operational backup 
  # only, which continuously backs up data and does not need a schedule
}
