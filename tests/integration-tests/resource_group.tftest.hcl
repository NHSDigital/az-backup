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
    backup_vault_name       = run.setup_tests.backup_vault_name
    tags                    = run.setup_tests.tags
  }

  assert {
    condition     = azurerm_resource_group.resource_group.name == var.resource_group_name
    error_message = "Resource group name not as expected."
  }

  assert {
    condition     = azurerm_resource_group.resource_group.location == var.resource_group_location
    error_message = "Resource group location not as expected."
  }

  assert {
    condition     = length(azurerm_resource_group.resource_group.tags) == length(run.setup_tests.tags)
    error_message = "Tags not as expected."
  }

  assert {
    condition = alltrue([
      for tag_key, tag_value in run.setup_tests.tags :
      lookup(azurerm_resource_group.resource_group.tags, tag_key, null) == tag_value
    ])
    error_message = "Tags not as expected."
  }
}
