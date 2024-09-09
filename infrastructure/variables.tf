variable "use_local_backend" {
  description = "Set to true to use the local backend instead of azurerm"
  type        = bool
  default     = false
}

variable "vault_name" {
  type    = string
  default = "myvault"
}

variable "vault_location" {
  type    = string
  default = "UK South"
}

variable "vault_redundancy" {
  type    = string
  default = "LocallyRedundant"
}
