resource "azurerm_data_protection_backup_policy_blob_storage" "backup_policy" {
  name                             = coalesce(length(trimspace(var.backup_policy_name_override)) > 0 ? var.backup_policy_name_override : null, "bkpol-blob-${var.backup_name}")
  vault_id                         = var.vault.id
  vault_default_retention_duration = var.retention_period
  backup_repeating_time_intervals  = var.backup_intervals
  time_zone                        = var.time_zone

  dynamic "retention_rule" {
    for_each = var.enable_daily_retention_rule ? [1] : []
    content {
      name     = "daily-retention"
      priority = 9

      criteria {
        absolute_criteria = "AllBackup"
      }

      life_cycle {
        data_store_type = "VaultStore"
        duration        = var.retention_period
      }
    }
  }
}
