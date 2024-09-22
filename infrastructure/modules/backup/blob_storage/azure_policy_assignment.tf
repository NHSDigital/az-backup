resource "azurerm_subscription_policy_assignment" "create_backup_instance" {
  name                 = "policyass-${var.vault_name}-create-backup-instance-blob-storage"
  policy_definition_id = azurerm_policy_definition.create_backup_instance.id
  subscription_id      = var.subscription_id
  location             = var.vault_location
  identity {
    type = "SystemAssigned"
  }
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

# TODO: check if this is needed, remove if definately not 
# data "azurerm_role_definition" "contributor" {
#   name  = "Contributor"
#   scope = var.subscription_id
# }

# resource "azurerm_role_assignment" "subscription_contributor" {
#   scope              = var.subscription_id
#   role_definition_id = data.azurerm_role_definition.contributor.id
#   principal_id       = azurerm_subscription_policy_assignment.create_backup_instance.identity[0].principal_id
# }
