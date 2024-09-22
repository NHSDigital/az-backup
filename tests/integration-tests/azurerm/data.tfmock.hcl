mock_data "azurerm_subscription" {
  defaults = {
    id = "/subscriptions/12345678-1234-9876-4563-123456789012"
  }
}

mock_resource "azurerm_data_protection_backup_vault" {
  defaults = {
    id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.DataProtection/backupVaults/bvault-testvault"
  }
}

mock_resource "azurerm_policy_definition" {
  defaults = {
    id = "/subscriptions/12345678-1234-9876-4563-123456789012/providers/Microsoft.Authorization/policyDefinitions/policy"
  }
}