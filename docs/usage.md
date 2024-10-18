# Usage

## Overview

To use the az-backup terraform module, create a terraform module in your own code and set the source as the az-backup repository.

[See the following link for more information about using github as the source of a terraform module.](https://developer.hashicorp.com/terraform/language/modules/sources#github)

The az-backup module resides in the `./infrastructure` sub directory of the repository, so you need to specify that in the module source by using the double-slash syntax [as explained in this guide](https://developer.hashicorp.com/terraform/language/modules/sources#modules-in-package-sub-directories).

In future we will use release tags to ensure consumers can depend on a specific release of the module, however this has not currently been implemented.

The module will create a dedicated resource group to contain the backup vault, therefore the resource group name provided to the module must be unique within the scope of the subscription.

## Example

The following is an example of how the module should be used:

```terraform
module "my_backup" {
  source                     = "github.com/nhsdigital/az-backup//infrastructure"
  resource_group_name        = "rg-mybackup"
  resource_group_location    = "uksouth"
  backup_vault_name          = "bvault-mybackup"
  backup_vault_redundancy    = "LocallyRedundant"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.my_workspace.id
  tags = {
    tagOne   = "tagOneValue"
    tagTwo   = "tagTwoValue"
    tagThree = "tagThreeValue"
  }
  blob_storage_backups = {
    backup1 = {
      backup_name        = "storage1"
      retention_period   = "P7D"
      storage_account_id = azurerm_storage_account.my_storage_account_1.id
    }
    backup2 = {
      backup_name        = "storage2"
      retention_period   = "P30D"
      storage_account_id = azurerm_storage_account.my_storage_account_2.id
    }
  }
  managed_disk_backups = {
    backup1 = {
      backup_name      = "disk1"
      retention_period = "P7D"
      backup_intervals = ["R/2024-01-01T00:00:00+00:00/P1D"]
      managed_disk_id  = azurerm_managed_disk.my_managed_disk_1.id
      managed_disk_resource_group = {
        id   = azurerm_resource_group.my_resource_group.id
        name = azurerm_resource_group.my_resource_group.name
      }
    }
    backup2 = {
      backup_name      = "disk2"
      retention_period = "P30D"
      backup_intervals = ["R/2024-01-01T00:00:00+00:00/P2D"]
      managed_disk_id  = azurerm_managed_disk.my_managed_disk_2.id
      managed_disk_resource_group = {
        id   = azurerm_resource_group.my_resource_group.id
        name = azurerm_resource_group.my_resource_group.name
      }
    }
  }
  postgresql_flexible_server_backups = {
    backup1 = {
      backup_name      = "server1"
      retention_period = "P7D"
      backup_intervals = ["R/2024-01-01T00:00:00+00:00/P1D"]
      server_id  = azurerm_postgresql_flexible_server.my_server_1.id
      server_resource_group_id = azurerm_resource_group.my_resource_group.id
    }
    backup2 = {
      backup_name      = "server2"
      retention_period = "P30D"
      backup_intervals = ["R/2024-01-01T00:00:00+00:00/P2D"]
      server_id  = azurerm_postgresql_flexible_server.my_server_2.id
      server_resource_group_id = azurerm_resource_group.my_resource_group.id
    }
  }
}
```

## Deployment Identity

To deploy the module an Azure identity (typically an app registration with client secret) is required which has been assigned the following roles at the subscription level:

* Contributor (required to create resources)
* Role Based Access Control Administrator (to assign roles to the backup vault managed identity) **with a condition that limits the roles which can be assigned to:**
    * Disk Backup Reader
    * Disk Snapshot Contributor
    * PostgreSQL Flexible Server Long Term Retention Backup Role
    * Storage Account Backup Contributor
    * Reader

## Module Variables

| Name | Description | Mandatory | Default |
|------|-------------|-----------|---------|
| `resource_group_name` | The name of the resource group that is created to contain the vault - this cannot be an existing resource group. | Yes | n/a |
| `resource_group_location` | The location of the resource group that is created to contain the vault. | No | `uksouth` |
| `backup_vault_name` | The name of the backup vault. The value supplied will be automatically prefixed with `rg-nhsbackup-`. If more than one az-backup module is created, this value must be unique across them. | Yes | n/a |
| `backup_vault_redundancy` | The redundancy of the vault, e.g. `GeoRedundant`. [See the following link for the possible values](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/data_protection_backup_vault#redundancy) | No | `LocallyRedundant` |
| `log_analytics_workspace_id` | The id of the log analytics workspace that backup telemetry and diagnostics should be sent to. When no value is provided then diagnostics will not be sent anywhere. | No | n/a |
| `tags` | A map of tags which will be applied to the resource group and backup vault. When no tags are specified then no tags are added. | No | n/a |
| `blob_storage_backups` | A map of blob storage backups that should be created. For each backup the following values should be provided: `storage_account_id`, `backup_name` and `retention_period`. When no value is provided then no backups are created. | No | n/a |
| `blob_storage_backups.storage_account_id` | The id of the storage account that should be backed up. | Yes | n/a |
| `blob_storage_backups.backup_name` | The name of the backup, which must be unique across blob storage backups. | Yes | n/a |
| `blob_storage_backups.retention_period` | How long the backed up data will be retained for, which should be in `ISO 8601` duration format. [See the following link for the possible values](https://en.wikipedia.org/wiki/ISO_8601#Durations). | Yes | n/a |
| `managed_disk_backups` | A map of managed disk backups that should be created. For each backup the following values should be provided: `managed_disk_id`, `backup_name` and `retention_period`. When no value is provided then no backups are created. | No | n/a |
| `managed_disk_backups.managed_disk_id` | The id of the managed disk that should be backed up. | Yes | n/a |
| `managed_disk_backups.backup_name` | The name of the backup, which must be unique across managed disk backups. | Yes | n/a |
| `managed_disk_backups.retention_period` | How long the backed up data will be retained for, which should be in `ISO 8601` duration format. [See the following link for the possible values](https://en.wikipedia.org/wiki/ISO_8601#Durations). | Yes | n/a |
| `managed_disk_backups.backup_intervals` | A list of intervals at which backups should be taken, which should be in `ISO 8601` duration format. [See the following link for the possible values](https://en.wikipedia.org/wiki/ISO_8601#Time_intervals). | Yes | n/a |
| `postgresql_flexible_server_backups` | A map of postgresql flexible server backups that should be created. For each backup the following values should be provided: `backup_name`, `server_id`, `server_resource_group_id`, `retention_period` and `backup_intervals`. When no value is provided then no backups are created. | No | n/a |
| `postgresql_flexible_server_backups.backup_name` | The name of the backup, which must be unique across postgresql flexible server backups. | Yes | n/a |
| `postgresql_flexible_server_backups.server_id` | The id of the postgresql flexible server that should be backed up. | Yes | n/a |
| `postgresql_flexible_server_backups.server_resource_group_id` | The id of the resource group which the postgresql flexible server resides in. | Yes | n/a |
| `postgresql_flexible_server_backups.retention_period` | How long the backed up data will be retained for, which should be in `ISO 8601` duration format. [See the following link for the possible values](https://en.wikipedia.org/wiki/ISO_8601#Durations). | Yes | n/a |
| `postgresql_flexible_server_backups.backup_intervals` | A list of intervals at which backups should be taken, which should be in `ISO 8601` duration format. [See the following link for the possible values](https://en.wikipedia.org/wiki/ISO_8601#Time_intervals). | Yes | n/a |
