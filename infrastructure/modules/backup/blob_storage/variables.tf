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
