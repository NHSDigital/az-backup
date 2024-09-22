resource "azurerm_subscription_policy_assignment" "create_backup_instance" {
  name                 = "policyass-${var.vault_name}-create-backup-instance-blob-storage"
  policy_definition_id = azurerm_policy_definition.create_backup_instance.id
  subscription_id      = var.subscription_id
  parameters = jsonencode({
    vaultName = {
      value = var.vault_name
    }
    backupInstanceName = {
      value = "bkinst-${var.vault_name}-${var.backup_name}"
    }
    backupPolicyId = {
      value = azurerm_data_protection_backup_policy_blob_storage.backup_policy.id
    }
  })
}
