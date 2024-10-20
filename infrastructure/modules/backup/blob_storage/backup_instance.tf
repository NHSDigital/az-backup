resource "azurerm_role_assignment" "role_assignment" {
  scope                = var.storage_account_id
  role_definition_name = "Storage Account Backup Contributor"
  principal_id         = var.vault.identity[0].principal_id
}

resource "azurerm_data_protection_backup_instance_blob_storage" "backup_instance" {
  name                            = "bkinst-blob-${var.backup_name}"
  vault_id                        = var.vault.id
  location                        = var.vault.location
  storage_account_id              = var.storage_account_id
  backup_policy_id                = azurerm_data_protection_backup_policy_blob_storage.backup_policy.id
  storage_account_container_names = var.storage_account_containers

  depends_on = [
    azurerm_role_assignment.role_assignment
  ]
}
