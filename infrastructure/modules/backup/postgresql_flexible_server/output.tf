output "backup_policy" {
  value = azurerm_data_protection_backup_policy_postgresql_flexible_server.backup_policy
}

output "backup_instance" {
  value = azurerm_data_protection_backup_instance_postgresql_flexible_server.backup_instance
}