resource "azurerm_data_protection_backup_policy_disk" "backup_policy" {
  name                            = "bkpol-disk-${var.backup_name}"
  vault_id                        = var.vault.id
  default_retention_duration      = var.retention_period
  backup_repeating_time_intervals = var.backup_intervals
}
