variable "vault_id" {
  type = string
}

variable "vault_name" {
  type = string
}

variable "vault_location" {
  type = string
}

variable "vault_principal_id" {
  type = string
}

variable "backup_name" {
  type = string
}

variable "retention_period" {
  type = string
}

variable "backup_intervals" {
  type = list(string)
}

variable "managed_disk_id" {
  type = string
}

variable "managed_disk_resource_group" {
  type = object({
    id   = string
    name = string
  })
}

variable "assign_resource_group_level_roles" {
  type = bool
}
