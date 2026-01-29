resource "azurerm_data_protection_backup_policy_blob_storage" "backup_policy" {
  name                             = coalesce(length(trimspace(var.backup_policy_name_override)) > 0 ? var.backup_policy_name_override : null, "bkpol-blob-${var.backup_name}")
  vault_id                         = var.vault.id
  vault_default_retention_duration = var.retention_period
  backup_repeating_time_intervals  = var.backup_intervals
}
