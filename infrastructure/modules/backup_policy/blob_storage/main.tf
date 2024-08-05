resource "azurerm_data_protection_backup_policy_blob_storage" "backup_policy" {
  name                                   = var.policy_name
  vault_id                               = var.vault_id
  operational_default_retention_duration = var.retention_period
}
