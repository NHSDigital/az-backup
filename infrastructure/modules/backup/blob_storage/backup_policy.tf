resource "azurerm_data_protection_backup_policy_blob_storage" "backup_policy" {
  name                             = "bkpol-blob-${var.backup_name}"
  vault_id                         = var.vault.id
  vault_default_retention_duration = var.retention_period
  backup_repeating_time_intervals  = var.backup_intervals
}
