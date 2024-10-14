terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }
  }
}

resource "random_pet" "vault_name" {
  length = 4
}

output "vault_name" {
  value = random_pet.vault_name.id
}

output "tags" {
  value = {
    environment         = "production"
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
