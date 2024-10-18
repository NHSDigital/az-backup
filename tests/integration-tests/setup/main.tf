terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }
  }
}

resource "random_pet" "backup_vault_name" {
  length = 4
}

output "resource_group_name" {
  value = "rg-${random_pet.backup_vault_name.id}"
}

output "backup_vault_name" {
  value = "bvault-${random_pet.backup_vault_name.id}"
}

output "tags" {
  value = {
    tagOne   = "tagOneValue"
    tagTwo   = "tagTwoValue"
    tagThree = "tagThreeValue"
  }
}
