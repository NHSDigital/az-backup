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

variable "storage_account_id" {
  type = string
}

variable "storage_account_containers" {
  type = list(string)
}

variable "backup_policy_name_override" {
  type    = string
  default = null
}

variable "backup_instance_name_override" {
  type    = string
  default = null
}
