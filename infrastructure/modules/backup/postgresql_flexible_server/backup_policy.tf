resource "azurerm_data_protection_backup_policy_postgresql_flexible_server" "backup_policy" {
  name                            = "bkpol-pgflex-${var.backup_name}"
  vault_id                        = var.vault.id
  backup_repeating_time_intervals = var.backup_intervals

  default_retention_rule {
    life_cycle {
      duration        = var.retention_period
      data_store_type = "VaultStore"
    }
  }
}
