run "setup_tests" {
    module {
        source = "./tests/setup"
    }
}

run "create_bucket" {
  command = plan

  variables {
    vault_name = "${run.setup_tests.vault_name}"
  }

  # Check that the vault name is correct
  assert {
    condition     = azurerm_data_protection_backup_vault.backup_vault.name == "bvault-${run.setup_tests.vault_name}"
    error_message = "Invalid backup vault name"
  }
}