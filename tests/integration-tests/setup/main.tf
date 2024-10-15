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
    environment     = "production"
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
