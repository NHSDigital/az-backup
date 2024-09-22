resource "azurerm_subscription_policy_assignment" "create_backup_instance" {
  name                 = "policyass-${var.vault_name}-backup-blob-storage"
  display_name         = "[AZ-BACKUP] Configure backup on blob storage accounts with a given tag"
  policy_definition_id = azurerm_policy_definition.create_backup_instance.id
  subscription_id      = var.subscription_id
  location             = var.vault_location
  identity {
    type = "SystemAssigned"
  }
  parameters = jsonencode({
    backupVaultId = {
      value = var.vault_id
    }
    backupPolicyId = {
      value = azurerm_data_protection_backup_policy_blob_storage.backup_policy.id
    }
    backupInstanceName = {
      value = "bkinst-${var.vault_name}-${var.backup_name}"
    }
  })
}

resource "azurerm_role_assignment" "backup_operator" {
  scope                = var.subscription_id
  role_definition_name = "Backup Operator"
  principal_id         = azurerm_subscription_policy_assignment.create_backup_instance.identity[0].principal_id
}

resource "azurerm_role_assignment" "role_based_access_control_administrator" {
  scope                = var.subscription_id
  role_definition_name = "Role Based Access Control Administrator"
  principal_id         = azurerm_subscription_policy_assignment.create_backup_instance.identity[0].principal_id
  condition_version    = "2.0"

  # This condition restricts write/delete of a role assignment to the specified roles only:
  # e5e2a7ff-d759-4cd2-bb51-3152d37e2eb1 = "Storage Account Backup Operator" 
  condition = <<-EOT
  (
    (
      !(ActionMatches{'Microsoft.Authorization/roleAssignments/write'})
    )
    OR 
    (
      @Request[Microsoft.Authorization/roleAssignments:RoleDefinitionId] ForAnyOfAnyValues:GuidEquals {e5e2a7ff-d759-4cd2-bb51-3152d37e2eb1}
    )
  )
  AND
  (
    (
      !(ActionMatches{'Microsoft.Authorization/roleAssignments/delete'})
    )
    OR 
    (
      @Resource[Microsoft.Authorization/roleAssignments:RoleDefinitionId] ForAnyOfAnyValues:GuidEquals {e5e2a7ff-d759-4cd2-bb51-3152d37e2eb1}
    )
  )
  EOT
}
