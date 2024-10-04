output "backup_policy" {
  value = azurerm_data_protection_backup_policy_disk.backup_policy
}

output "backup_instance" {
  value = azurerm_data_protection_backup_instance_disk.backup_instance
}