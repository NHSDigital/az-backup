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

variable "backup_policy_naming_template" {
  type    = string
  default = "{resource_abbreviation}-{resource_type}-{backup_name}"
}

variable "backup_instance_naming_template" {
  type    = string
  default = "{resource_abbreviation}-{resource_type}-{backup_name}"
}

variable "time_zone" {
  type    = string
  default = null
}

variable "enable_daily_retention_rule" {
  type    = bool
  default = false
}
