variable "policy_name" {
  type = string
}

variable "vault_name" {
  type = string
}

variable "resource_group_name" {
  type = string
}

variable "retention_period" {
  type = string
}

variable "backup_intervals" {
  type = list(string)
}
