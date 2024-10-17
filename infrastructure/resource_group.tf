resource "azurerm_resource_group" "resource_group" {
  location = var.resource_group_location
  name     = var.resource_group_name
}
