mock_provider "azurerm" {
  source = "./azurerm"
}

run "setup_tests" {
  module {
    source = "./setup"
  }
}

run "create_resource_group" {
  command = apply

  module {
    source = "../../infrastructure"
  }

  variables {
    vault_name     = run.setup_tests.vault_name
    vault_location = "uksouth"
  }

  # Check that the name is as expected
  assert {
    condition     = azurerm_resource_group.resource_group.name == "rg-nhsbackup-${var.vault_name}"
    error_message = "Resource group name not as expected."
  }

  # Check that the location is as expected
  assert {
    condition     = azurerm_resource_group.resource_group.location == var.vault_location
    error_message = "Resource group location not as expected."
  }
}