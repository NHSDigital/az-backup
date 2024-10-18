variable "resource_group_name" {
  description = "The name of the resource group which the backup vault will be created in - must be unique within the subscription"
  type        = string
}

variable "resource_group_location" {
  description = "The location of the resource group which the backup vault will be created in"
  type        = string
  default     = "uksouth"
}

variable "backup_vault_name" {
  description = "The name of the backup vault"
  type        = string
}

variable "backup_vault_redundancy" {
  description = "The redundancy of the backup vault"
  type        = string
  default     = "LocallyRedundant"
}

variable "log_analytics_workspace_id" {
  description = "The id of the log analytics workspace to use for backup vault diagnostic settings"
  type        = string
  default     = ""
}

variable "tags" {
  description = "A map of tags to assign to the resource group, as mandated by the CCOE tagging policy"
  type        = map(string)
  default     = {}
}

variable "blob_storage_backups" {
  description = "A map of blob storage backups to create"
  type = map(object({
    backup_name        = string
    retention_period   = string
    storage_account_id = string
  }))
  default = {}
}

variable "managed_disk_backups" {
  description = "A map of managed disk backups to create"
  type = map(object({
    backup_name      = string
    retention_period = string
    backup_intervals = list(string)
    managed_disk_id  = string
    managed_disk_resource_group = object({
      id   = string
      name = string
    })
  }))
  default = {}
}

variable "postgresql_flexible_server_backups" {
  description = "A map of postgresql flexible server backups to create"
  type = map(object({
    backup_name              = string
    retention_period         = string
    backup_intervals         = list(string)
    server_id                = string
    server_resource_group_id = string
  }))
  default = {}
}

