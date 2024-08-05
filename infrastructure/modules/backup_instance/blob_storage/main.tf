resource "azurerm_role_assignment" "role_assignment" {
  scope                = var.storage_account_id
  role_definition_name = "Storage Account Backup Contributor"
  principal_id         = var.vault_principal_id
}

resource "azurerm_data_protection_backup_instance_blob_storage" "backup_instance" {
  name               = var.instance_name
  vault_id           = var.vault_id
  location           = var.vault_location
  storage_account_id = var.storage_account_id
  backup_policy_id   = var.policy_id

  depends_on = [
    azurerm_role_assignment.role_assignment
  ]
}
