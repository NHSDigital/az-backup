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
  description = "A map of tags to assign to the resources created by the module"
  type        = map(string)
  default     = {}
}

variable "use_extended_retention" {
  description = "A flag which allows the use of extended retention periods beyond 7 days"
  type        = bool
  default     = false
}

locals {
  # The valid backup retention period - up to 7 days, which can be extended when use_extended_retention is set to true
  valid_retention_periods = [for days in range(1, 8) : "P${days}D"]
}

variable "blob_storage_backups" {
  description = "A map of blob storage backups to create"
  type = map(object({
    backup_name                = string
    retention_period           = string
    backup_intervals           = list(string)
    storage_account_id         = string
    storage_account_containers = list(string)
  }))

  default = {}

  validation {
    condition     = alltrue([for k, v in var.blob_storage_backups : length(v.backup_intervals) > 0])
    error_message = "At least one backup interval must be provided."
  }

  validation {
    condition     = alltrue([for k, v in var.blob_storage_backups : length(v.storage_account_containers) > 0])
    error_message = "At least one storage account container must be provided."
  }

  validation {
    condition     = var.use_extended_retention == true || alltrue([for k, v in var.blob_storage_backups : contains(local.valid_retention_periods, v.retention_period)])
    error_message = "Invalid retention period: valid periods are up to 7 days. If you require a longer retention period then please set use_exetended_retention to true."
  }
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

  validation {
    condition     = alltrue([for k, v in var.managed_disk_backups : length(v.backup_intervals) > 0])
    error_message = "At least one backup interval must be provided."
  }

  validation {
    condition     = var.use_extended_retention == true || alltrue([for k, v in var.managed_disk_backups : contains(local.valid_retention_periods, v.retention_period)])
    error_message = "Invalid retention period: valid periods are up to 7 days. If you require a longer retention period then please set use_exetended_retention to true."
  }
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

  validation {
    condition     = alltrue([for k, v in var.postgresql_flexible_server_backups : length(v.backup_intervals) > 0])
    error_message = "At least one backup interval must be provided."
  }

  validation {
    condition     = var.use_extended_retention == true || alltrue([for k, v in var.postgresql_flexible_server_backups : contains(local.valid_retention_periods, v.retention_period)])
    error_message = "Invalid retention period: valid periods are up to 7 days. If you require a longer retention period then please set use_exetended_retention to true."
  }
}