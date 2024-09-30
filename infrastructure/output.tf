output "vault_name" {
  value = azurerm_data_protection_backup_vault.backup_vault.name
}

output "vault_location" {
  value = azurerm_data_protection_backup_vault.backup_vault.location
}

output "vault_redundancy" {
  value = azurerm_data_protection_backup_vault.backup_vault.redundancy
}
