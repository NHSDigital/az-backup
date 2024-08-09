resource "azurerm_data_protection_backup_policy_kubernetes_cluster" "backup_policy" {
  name                            = var.policy_name
  vault_name                      = var.vault_name
  resource_group_name             = var.resource_group_name
  backup_repeating_time_intervals = var.backup_intervals
  default_retention_rule {
    life_cycle {
      duration        = var.retention_period
      data_store_type = "OperationalStore"
    }
  }
}
