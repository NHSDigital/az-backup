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

variable "blob_storage_backups" {
  type = map(object({
    backup_name        = string
    retention_period   = string
    storage_account_id = string
  }))
  default = {}
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
}

variable "postgresql_flexible_server_backups" {
  type = map(object({
    backup_name              = string
    retention_period         = string
    backup_intervals         = list(string)
    server_id                = string
    server_resource_group_id = string
  }))
  default = {}
}

variable "vault_immutabilitySettings" {
  type    = string
  default = "Disabled" 	#"Disabled" "Locked" "Unlocked"
}