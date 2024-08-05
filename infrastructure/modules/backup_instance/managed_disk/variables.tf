variable "instance_name" {
  type = string
}

variable "vault_id" {
  type = string
}

variable "vault_location" {
  type = string
}

variable "vault_principal_id" {
  type = string
}

variable "policy_id" {
  type = string
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
