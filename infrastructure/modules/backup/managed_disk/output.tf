output "id" {
  value = azurerm_data_protection_backup_policy_disk.backup_policy.id
}

output "name" {
  value = azurerm_data_protection_backup_policy_disk.backup_policy.name
}

output "vault_id" {
  value = azurerm_data_protection_backup_policy_disk.backup_policy.vault_id
}

output "retention_period" {
  value = azurerm_data_protection_backup_policy_disk.backup_policy.default_retention_duration
}

output "backup_intervals" {
  value = azurerm_data_protection_backup_policy_disk.backup_policy.backup_repeating_time_intervals
}
