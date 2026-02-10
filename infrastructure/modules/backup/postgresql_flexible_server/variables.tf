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

variable "server_id" {
  type = string
}

variable "server_resource_group_id" {
  type = string
}

variable "assign_resource_group_level_roles" {
  type = bool
}

variable "backup_policy_naming_template" {
  type    = string
  default = "{resource_abbreviation}-{resource_type}-{backup_name}"
}

variable "backup_instance_naming_template" {
  type    = string
  default = "{resource_abbreviation}-{resource_type}-{backup_name}"
}
