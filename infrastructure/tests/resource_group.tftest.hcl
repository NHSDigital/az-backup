run "setup_tests" {
  module {
    source = "./tests/setup"
  }
}

run "create_bucket" {
  command = plan

  variables {
    vault_name = run.setup_tests.vault_name
  }

  # Check that the resource group name is as expected
  assert {
    condition     = azurerm_resource_group.resource_group.name == "rg-nhsbackup-${run.setup_tests.vault_name}"
    error_message = "Resource group name not as expected."
  }
}