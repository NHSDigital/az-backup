variable "policy_name" {
  type = string
}

variable "vault_id" {
  type = string
}

variable "retention_period" {
  type = string
}

variable "backup_intervals" {
  type = list(string)
}
