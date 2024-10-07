variable "vault_name" {
  type = string
}

variable "vault_location" {
  type    = string
  default = "uksouth"
}

variable "datastore_type" {
  type    = string
  default = "VaultStore"
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
variable "monitoringSettings" {
  type    = string
  default = "Disabled"
}

variable "immutabilitySettings" {
  type    = string
  default = "Unlocked"
}

variable "softDeleteSettingsretentionDays" {
  type    = string
  default = "14"
}

variable "softDeleteSettingsState" {
  type    = string
  default = "Off"
}

variable "tags"{
    type = map

}