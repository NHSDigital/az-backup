variable "vault_name" {
  description = "The name of the vault"
  type        = string
}

variable "vault_location" {
  description = "The location of the vault"
  type        = string
  default     = "uksouth"
}

variable "vault_redundancy" {
  description = "The redundancy of the vault"
  type        = string
  default     = "LocallyRedundant"
}

variable "tags" {
  description = "A map of tags to assign to the resource group, as mandated by the CCOE tagging policy"
  type = object({
    environment         = string
    owner               = string
    created_by          = string
    costing_pcode       = string
    ch_cost_centre      = string
    project             = string
    service_level       = string
    directorate         = string
    sub_directorate     = string
    data_classification = string
    service_product     = string
    team                = string
  })

  validation {
    condition     = contains(["production", "development", "integration", "staging"], var.tags["environment"])
    error_message = "The environment tag must be one of the following values: production, development, integration, staging."
  }

  validation {
    condition     = contains(["bronze", "silver", "gold", "platinum"], var.tags["service_level"])
    error_message = "The service_level tag must be one of the following values: bronze, silver, gold, platinum."
  }

  validation {
    condition     = contains(["1", "2", "3", "4", "5"], var.tags["data_classification"])
    error_message = "The data_classification tag must be one of the following values: 1, 2, 3, 4, 5."
  }
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
