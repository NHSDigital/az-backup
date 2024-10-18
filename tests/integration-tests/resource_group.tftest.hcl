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
    tags           = run.setup_tests.tags
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

run "validate_resource_group_tags_environment" {
  command = plan

  module {
    source = "../../infrastructure"
  }

  variables {
    vault_name     = run.setup_tests.vault_name
    vault_location = "uksouth"
    tags = {
      environment     = "invalid-environment"
      cost_code       = "code_value"
      created_by      = "creator_name"
      created_date    = "01/01/2024"
      tech_lead       = "tech_lead_name"
      requested_by    = "requester_name"
      service_product = "product_name"
      team            = "team_name"
      service_level   = "gold"
    }
  }

  expect_failures = [
    var.tags,
  ]
}

run "validate_resource_group_tags_service_level" {
  command = plan

  module {
    source = "../../infrastructure"
  }

  variables {
    vault_name     = run.setup_tests.vault_name
    vault_location = "uksouth"
    tags = {
      environment     = "production"
      cost_code       = "code_value"
      created_by      = "creator_name"
      created_date    = "01/01/2024"
      tech_lead       = "tech_lead_name"
      requested_by    = "requester_name"
      service_product = "product_name"
      team            = "team_name"
      service_level   = "invalid-service-level"
    }
  }

  expect_failures = [
    var.tags,
  ]
}
