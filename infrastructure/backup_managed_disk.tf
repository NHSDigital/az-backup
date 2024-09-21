module "managed_disk_backup" {
  source           = "./modules/backup/managed_disk"
  policy_name      = "bkpol-${var.vault_name}-manageddisk"
  vault_id         = azurerm_data_protection_backup_vault.backup_vault.id
  retention_period = "P7D"                               # 7 days
  backup_intervals = ["R/2024-01-01T00:00:00+00:00/P1D"] # Once per day at 00:00
}
