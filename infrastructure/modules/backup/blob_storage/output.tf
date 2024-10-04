output "backup_policy" {
  value = azurerm_data_protection_backup_policy_blob_storage.backup_policy
}

output "backup_instance" {
  value = azurerm_data_protection_backup_instance_blob_storage.backup_instance
}