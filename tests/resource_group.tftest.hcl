run "create_resource_group" {
  command = plan

  module {
    source = "../infrastructure"
  }

  variables {
    vault_name     = "testvault"
    vault_location = "uksouth"
  }

  # Check that the name is as expected
  assert {
    condition     = azurerm_resource_group.resource_group.name == "rg-nhsbackup-testvault"
    error_message = "Resource group name not as expected."
  }

  # Check that the location is as expected
  assert {
    condition     = azurerm_resource_group.resource_group.location == "uksouth"
    error_message = "Resource group location not as expected."
  }
}