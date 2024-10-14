mock_provider "azurerm" {
  source = "./azurerm"
}

run "setup_tests" {
  module {
    source = "./setup"
  }
}

run "create_backup_vault" {
  command = apply

  module {
    source = "../../infrastructure"
  }

  variables {
    vault_name       = run.setup_tests.vault_name
    vault_location   = "uksouth"
    vault_redundancy = "LocallyRedundant"
    tags             = run.setup_tests.tags
  }

  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.name == "bvault-${var.vault_name}"
    error_message = "Backup vault name not as expected."
  }

  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.resource_group_name == azurerm_resource_group.resource_group.name
    error_message = "Resource group not as expected."
  }

  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.location == var.vault_location
    error_message = "Backup vault location not as expected."
  }

  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.datastore_type == "VaultStore"
    error_message = "Backup vault datastore type not as expected."
  }

  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.redundancy == var.vault_redundancy
    error_message = "Backup vault redundancy not as expected."
  }

  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.soft_delete == "Off"
    error_message = "Backup vault soft delete not as expected."
  }

  assert {
    condition     = length(azurerm_data_protection_backup_vault.backup_vault.identity[0].principal_id) > 0
    error_message = "Backup vault identity not as expected."
  }
}