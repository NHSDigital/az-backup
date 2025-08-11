terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
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

output "log_analytics_workspace_id" {
  value = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.OperationalInsights/workspaces/workspace1"
}

output "tags" {
  value = {
    tagOne   = "tagOneValue"
    tagTwo   = "tagTwoValue"
    tagThree = "tagThreeValue"
  }
}
