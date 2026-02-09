resource "azurerm_data_protection_backup_policy_blob_storage" "backup_policy" {
  name                             = local.backup_policy_name
  vault_id                         = var.vault.id
  vault_default_retention_duration = var.retention_period
  backup_repeating_time_intervals  = var.backup_intervals
  time_zone                        = var.time_zone

  dynamic "retention_rule" {
    for_each = coalesce(var.enable_daily_retention_rule, false) ? [1] : []
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
