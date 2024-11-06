module "backup_vault" {
  source                     = "../infrastructure"
  resource_group_name        = "rg-nhsbackup-example-vault"
  resource_group_location    = "uksouth"
  backup_vault_name          = "bvault-nhsbackup-example"
  backup_vault_redundancy    = "LocallyRedundant"
  backup_vault_immutability  = "Disabled"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.log_analytics.id
  tags = {
    creator     = "Joe Bloggs"
    environment = "Development"
    fruit       = "Bananas"
  }
  blob_storage_backups = {
    backup1 = {
      backup_name        = "storage1"
      retention_period   = "P1D"
      backup_intervals   = ["R/2024-01-01T00:00:00+00:00/P1D"]
      storage_account_id = azurerm_storage_account.storage_account_one.id
      storage_account_containers = [
        azurerm_storage_container.storage_account_one_container_one.name,
        azurerm_storage_container.storage_account_one_container_two.name
      ]
    }
    backup2 = {
      backup_name        = "storage2"
      retention_period   = "P2D"
      backup_intervals   = ["R/2024-01-01T00:00:00+00:00/P2D"]
      storage_account_id = azurerm_storage_account.storage_account_two.id
      storage_account_containers = [
        azurerm_storage_container.storage_account_two_container_one.name,
        azurerm_storage_container.storage_account_two_container_two.name
      ]
    }
  }
  managed_disk_backups = {
    backup1 = {
      backup_name      = "disk1"
      retention_period = "P1D"
      backup_intervals = ["R/2024-01-01T00:00:00+00:00/P1D"]
      managed_disk_id  = azurerm_managed_disk.managed_disk.id
      managed_disk_resource_group = {
        id   = azurerm_resource_group.resource_group.id
        name = azurerm_resource_group.resource_group.name
      }
    }
  }
  postgresql_flexible_server_backups = {
    backup1 = {
      backup_name              = "server1"
      retention_period         = "P1D"
      backup_intervals         = ["R/2024-01-01T00:00:00+00:00/P1D"]
      server_id                = azurerm_postgresql_flexible_server.postgresql_flexible_server.id
      server_resource_group_id = azurerm_resource_group.resource_group.id
    }
  }
}