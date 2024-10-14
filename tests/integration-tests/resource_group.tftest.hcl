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
    tags           = run.setup_tests.tags
  }

  assert {
    condition     = azurerm_resource_group.resource_group.name == "rg-nhsbackup-${var.vault_name}"
    error_message = "Resource group name not as expected."
  }

  assert {
    condition     = azurerm_resource_group.resource_group.location == var.vault_location
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
      environment         = "invalid-environment"
      owner               = "owner_name"
      created_by          = "creator_name"
      costing_pcode       = "pcode_value"
      ch_cost_centre      = "cost_centre_value"
      project             = "project_name"
      service_level       = "gold"
      directorate         = "directorate_name"
      sub_directorate     = "sub_directorate_name"
      data_classification = "3"
      service_product     = "product_name"
      team                = "team_name"
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
      environment         = "production"
      owner               = "owner_name"
      created_by          = "creator_name"
      costing_pcode       = "pcode_value"
      ch_cost_centre      = "cost_centre_value"
      project             = "project_name"
      service_level       = "invalid-service-level"
      directorate         = "directorate_name"
      sub_directorate     = "sub_directorate_name"
      data_classification = "3"
      service_product     = "product_name"
      team                = "team_name"
    }
  }

  expect_failures = [
    var.tags,
  ]
}

run "validate_resource_group_tags_data_classification" {
  command = plan

  module {
    source = "../../infrastructure"
  }

  variables {
    vault_name     = run.setup_tests.vault_name
    vault_location = "uksouth"
    tags = {
      environment         = "production"
      owner               = "owner_name"
      created_by          = "creator_name"
      costing_pcode       = "pcode_value"
      ch_cost_centre      = "cost_centre_value"
      project             = "project_name"
      service_level       = "gold"
      directorate         = "directorate_name"
      sub_directorate     = "sub_directorate_name"
      data_classification = "invalid-data-classification"
      service_product     = "product_name"
      team                = "team_name"
    }
  }

  expect_failures = [
    var.tags,
  ]
}