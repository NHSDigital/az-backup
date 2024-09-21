output "backup_policy_id" {
  value = azurerm_data_protection_backup_policy_blob_storage.backup_policy.id
}

output "backup_policy_name" {
  value = azurerm_data_protection_backup_policy_blob_storage.backup_policy.name
}

output "vault_id" {
  value = azurerm_data_protection_backup_policy_blob_storage.backup_policy.vault_id
}

output "retention_period" {
  value = azurerm_data_protection_backup_policy_blob_storage.backup_policy.operational_default_retention_duration
}

output "azure_policy_definition_id" {
  value = azurerm_policy_definition.create_backup_instance.id
}

output "azure_policy_definition_name" {
  value = azurerm_policy_definition.create_backup_instance.name
}

output "azure_policy_definition_policy_type" {
  value = azurerm_policy_definition.create_backup_instance.policy_type
}

output "azure_policy_definition_mode" {
  value = azurerm_policy_definition.create_backup_instance.mode
}

output "azure_policy_assignment_id" {
  value = azurerm_subscription_policy_assignment.create_backup_instance.id
}

output "azure_policy_assignment_name" {
  value = azurerm_subscription_policy_assignment.create_backup_instance.name
}

output "azure_policy_assignment_subscription_id" {
  value = azurerm_subscription_policy_assignment.create_backup_instance.subscription_id
}

output "azure_policy_assignment_parameters" {
  value = azurerm_subscription_policy_assignment.create_backup_instance.parameters
}
