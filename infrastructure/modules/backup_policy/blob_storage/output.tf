output "id" {
  value = azurerm_data_protection_backup_policy_blob_storage.backup_policy.id
}

output "name" {
  value = azurerm_data_protection_backup_policy_blob_storage.backup_policy.name
}

output "vault_id" {
  value = azurerm_data_protection_backup_policy_blob_storage.backup_policy.vault_id
}

output "retention_period" {
  value = azurerm_data_protection_backup_policy_blob_storage.backup_policy.operational_default_retention_duration
}
