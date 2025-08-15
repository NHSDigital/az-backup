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

variable "create_role_assignment" {
  type        = bool
  description = "Whether to create the Storage Account Backup Contributor role assignment. Set to false for additional backups targeting the same storage account."
  default     = true
}
