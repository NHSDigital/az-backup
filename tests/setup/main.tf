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