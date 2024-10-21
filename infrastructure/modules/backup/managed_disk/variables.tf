variable "vault" {
  type = any
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
