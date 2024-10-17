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
    resource_group_name     = run.setup_tests.resource_group_name
    resource_group_location = "uksouth"
    backup_vault_name     = run.setup_tests.backup_vault_name
  }

  assert {
    condition     = azurerm_resource_group.resource_group.name == var.resource_group_name
    error_message = "Resource group name not as expected."
  }

  assert {
    condition     = azurerm_resource_group.resource_group.location == var.resource_group_location
    error_message = "Resource group location not as expected."
  }
}