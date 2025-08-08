# Usage

## Overview

To use the az-backup terraform module, create a terraform module in your own code and set the source as the az-backup repository.

[See the following link for more information about using github as the source of a terraform module.](https://developer.hashicorp.com/terraform/language/modules/sources#github)

The az-backup module resides in the `./infrastructure` sub directory of the repository, so you need to specify that in the module source by using the double-slash syntax [as explained in this guide](https://developer.hashicorp.com/terraform/language/modules/sources#modules-in-package-sub-directories).

By default, the module will create a dedicated resource group to contain the backup vault, therefore the resource group name provided to the module must be unique within the scope of the subscription. The creation of a dedicated resource group can be overridden if the vault needs to be deployed into an externally managed resource group.

## Immutability

Immutability is configured by setting the `backup_vault_immutability` variable. The variable can be set as `Disabled` (default), `Unlocked` and `Locked`.

**IMPORTANT:** A backup vault cannot be created in a `Locked` state, therefore you must first deploy it as `Unlocked`, and the update the configuration to `Locked` as a second step.

## Retention

By default the module restricts backup retention to 7 days, in order to protect against immutable copies of data being created which cannot be deleted.

To override the restriction set the `use_extended_retention` variable to true, which will allow you to set a retention of any length.

## Identity

To deploy the module an Azure identity (e.g. an app registration with client secret) is required which has been assigned the following roles at the subscription level:

* Contributor (to create resources)
* Role Based Access Control Administrator (to assign roles to the backup vault managed identity) **with a condition limiting the roles that can be assigned to:**
    * Disk Backup Reader
    * Disk Snapshot Contributor
    * PostgreSQL Flexible Server Long Term Retention Backup Role
    * Storage Account Backup Contributor
    * Reader

## Deployment

Configure the tenant, subscription and credentials of the identity as environment variables and deploy with terraform.

```pwsh
$env:ARM_TENANT_ID="<your-tenant-id>"
$env:ARM_SUBSCRIPTION_ID="<your-subscription-id>"
$env:ARM_CLIENT_ID="<your-client-id>"
$env:ARM_CLIENT_SECRET="<your-client-secret>"
```

## Example

The following is an example of how the module should be used - **update the ref with the release version that you want to use**:

```terraform
module "my_backup" {
  source                     = "github.com/nhsdigital/az-backup//infrastructure?ref=<version-number>"
  resource_group_name        = "rg-mybackup"
  resource_group_location    = "uksouth"
  create_resource_group      = true
  backup_vault_name          = "bvault-mybackup"
  backup_vault_redundancy    = "LocallyRedundant"
  backup_vault_immutability  = "Unlocked"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.my_workspace.id
  use_extended_retention     = true

  tags = {
    tagOne   = "tagOneValue"
    tagTwo   = "tagTwoValue"
    tagThree = "tagThreeValue"
  }
  
  blob_storage_backups = {
    backup1 = {
      backup_name                = "storage1"
      retention_period           = "P7D"
      backup_intervals           = ["R/2024-01-01T00:00:00+00:00/P1D"]
      storage_account_id         = azurerm_storage_account.my_storage_account_1.id
      storage_account_containers = ["container1", "container2"]
    }
    backup2 = {
      backup_name                = "storage2"
      retention_period           = "P30D"
      backup_intervals           = ["R/2024-01-01T00:00:00+00:00/P2D"]
      storage_account_id         = azurerm_storage_account.my_storage_account_2.id
      storage_account_containers = ["container1", "container2"]
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

### Input Variables

| Name | Description | Required | Default |
|------|-------------|-----------|---------|
| `resource_group_name` | The name of the resource group that is created to contain the vault - the resource group will be created if `create_resource_group` = true, and must be an existing resource group if `create_resource_group` = false. | Yes | n/a |
| `resource_group_location` | The location of the resource group. | No | `uksouth` |
| `create_resource_group` | States whether a resource group should be created. Setting this to `false` means the vault will be deployed into an externally managed resource group, the name of which is defined in `resource_group_name`. | No | `true` |
| `backup_vault_name` | The name of the backup vault. The value supplied will be automatically prefixed with `rg-nhsbackup-`. If more than one az-backup module is created, this value must be unique across them. | Yes | n/a |
| `backup_vault_redundancy` | The redundancy of the vault, e.g. `GeoRedundant`. [See the following link for the possible values.](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/data_protection_backup_vault#redundancy) | No | `LocallyRedundant` |
| `backup_vault_immutability` | The immutability of the vault, e.g. `Locked`. [See the following link for the possible values.](https://learn.microsoft.com/en-us/azure/templates/microsoft.dataprotection/backupvaults?pivots=deployment-language-terraform#immutabilitysettings-2) | No | `Disabled` |
| `log_analytics_workspace_id` | The id of the log analytics workspace that backup telemetry and diagnostics should be sent to. **NOTE** this variable was made mandatory in v2 of the module. | Yes | n/a |
| `tags` | A map of tags which will be applied to the resource group and backup vault. When no tags are specified then no tags are added. NOTE when using an externally managed resource group the tags will not be applied to it (they will still be applied to the backup vault). | No | n/a |
| `use_extended_retention` | If set to true, then the backup retention periods can be set to anything, otherwise they are limited to 7 days. | No | `false` |
| `blob_storage_backups` | A map of blob storage backups that should be created. For each backup the following values should be provided: `storage_account_id`, `backup_name` and `retention_period`. When no value is provided then no backups are created. | No | n/a |
| `blob_storage_backups.storage_account_id` | The id of the storage account that should be backed up. | Yes | n/a |
| `blob_storage_backups.storage_account_containers` | A list of containers in the storage account that should be backed up. | Yes | n/a |
| `blob_storage_backups.backup_name` | The name of the backup, which must be unique across blob storage backups. | Yes | n/a |
| `blob_storage_backups.retention_period` | How long the backed up data will be retained for, which should be in `ISO 8601` duration format. This must be specified in days, and can be up to 7 days unless `use_extended_retention` is on. [See the following link for more information about the format](https://en.wikipedia.org/wiki/ISO_8601#Durations). | Yes | n/a |
| `blob_storage_backups.backup_intervals` | A list of intervals at which backups should be taken, which should be in `ISO 8601` duration format. [See the following link for the possible values](https://en.wikipedia.org/wiki/ISO_8601#Time_intervals). | Yes | n/a |
| `managed_disk_backups` | A map of managed disk backups that should be created. For each backup the following values should be provided: `managed_disk_id`, `backup_name` and `retention_period`. When no value is provided then no backups are created. | No | n/a |
| `managed_disk_backups.managed_disk_id` | The id of the managed disk that should be backed up. | Yes | n/a |
| `managed_disk_backups.backup_name` | The name of the backup, which must be unique across managed disk backups. | Yes | n/a |
| `managed_disk_backups.retention_period` | How long the backed up data will be retained for, which should be in `ISO 8601` duration format. This must be specified in days, and can be up to 7 days unless `use_extended_retention` is on. [See the following link for more information about the format](https://en.wikipedia.org/wiki/ISO_8601#Durations). | Yes | n/a |
| `managed_disk_backups.backup_intervals` | A list of intervals at which backups should be taken, which should be in `ISO 8601` duration format. [See the following link for the possible values](https://en.wikipedia.org/wiki/ISO_8601#Time_intervals). | Yes | n/a |
| `postgresql_flexible_server_backups` | A map of postgresql flexible server backups that should be created. For each backup the following values should be provided: `backup_name`, `server_id`, `server_resource_group_id`, `retention_period` and `backup_intervals`. When no value is provided then no backups are created. | No | n/a |
| `postgresql_flexible_server_backups.backup_name` | The name of the backup, which must be unique across postgresql flexible server backups. | Yes | n/a |
| `postgresql_flexible_server_backups.server_id` | The id of the postgresql flexible server that should be backed up. | Yes | n/a |
| `postgresql_flexible_server_backups.server_resource_group_id` | The id of the resource group which the postgresql flexible server resides in. | Yes | n/a |
| `postgresql_flexible_server_backups.retention_period` | How long the backed up data will be retained for, which should be in `ISO 8601` duration format. This must be specified in days, and can be up to 7 days unless `use_extended_retention` is on. [See the following link for more information about the format](https://en.wikipedia.org/wiki/ISO_8601#Durations). | Yes | n/a |
| `postgresql_flexible_server_backups.backup_intervals` | A list of intervals at which backups should be taken, which should be in `ISO 8601` duration format. [See the following link for the possible values](https://en.wikipedia.org/wiki/ISO_8601#Time_intervals). | Yes | n/a |
