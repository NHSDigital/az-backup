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

variable "monitoringSettings" {
  type    = string
  default = "Disabled"
}

variable "immutabilitySettings" {
  type    = string
  default = "Unlocked"
}

variable "softDeleteSettingsretentionDays" {
  type    = string
  default = "14"
}

variable "softDeleteSettingsState" {
  type    = string
  default = "Off"
}

variable "tags"{
    type = map

}