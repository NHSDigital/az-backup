resource "azurerm_resource_group" "resource_group" {
  count    = var.create_resource_group ? 1 : 0
  location = var.resource_group_location
  name     = var.resource_group_name
  tags     = var.tags
}

data "azurerm_resource_group" "resource_group" {
  count = !var.create_resource_group ? 1 : 0
  name  = var.resource_group_name
}

locals {
  resource_group = var.create_resource_group ? azurerm_resource_group.resource_group[0] : data.azurerm_resource_group.resource_group[0]
}
