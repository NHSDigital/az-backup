mock_resource "azurerm_resource_group" {
  defaults = {
    id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group"
  }
}

mock_resource "azurerm_data_protection_backup_vault" {
  defaults = {
    id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.DataProtection/backupVaults/bvault-testvault"
  }
}

mock_resource "azurerm_data_protection_backup_policy_blob_storage" {
  defaults = {
    id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.DataProtection/backupVaults/bvault-testvault/backupPolicies/bkpol-testvault-testpolicy"
  }
}

mock_resource "azurerm_data_protection_backup_policy_disk" {
  defaults = {
    id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.DataProtection/backupVaults/bvault-testvault/backupPolicies/bkpol-testvault-testpolicy"
  }
}

mock_resource "azurerm_data_protection_backup_policy_postgresql_flexible_server" {
  defaults = {
    id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.DataProtection/backupVaults/bvault-testvault/backupPolicies/bkpol-testvault-testpolicy"
  }
}

mock_resource "azurerm_log_analytics_workspace" {
  defaults = {
    id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.OperationalInsights/workspaces/law-testworkspace"
  }  
}
