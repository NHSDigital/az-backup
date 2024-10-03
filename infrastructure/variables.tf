variable "vault_name" {
  type = string
}

variable "vault_location" {
  type    = string
  default = "uksouth"
}

variable "vault_redundancy" {
  type    = string
  default = "LocallyRedundant"
}

variable "use_extended_retention" {
  type    = bool
  default = false
}

locals {
  valid_retention_periods = (
    var.use_extended_retention
    ? [for days in range(1, 366) : "P${days}D"]
    : [for days in range(1, 8) : "P${days}D"]
  )
}

variable "blob_storage_backups" {
  type = map(object({
    backup_name        = string
    retention_period   = string
    storage_account_id = string
  }))

  default = {}

  validation {
    condition = alltrue([
      for k, v in var.blob_storage_backups :
      contains(local.valid_retention_periods, v.retention_period)
    ])
    error_message = "Invalid retention period. Valid periods are up to 7 days. If extended retention is enabled, valid periods are any duration less than 365 days (e.g., P30D, P60D, etc.)."
  }
}

variable "managed_disk_backups" {
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
    condition = alltrue([
      for k, v in var.managed_disk_backups :
      contains(local.valid_retention_periods, v.retention_period)
    ])
    error_message = "Invalid retention period. Valid periods are up to 7 days. If extended retention is enabled, valid periods are any duration less than 365 days (e.g., P30D, P60D, etc.)."
  }
}
