# The consumer of the module can configure whether the resource group should be created or
# not. If the resource group should be created, the module will create it. If the resource 
# group should not be created, the module will look up the existing resource group and 
# reference it as data. 
#
# The local variable resource_group means the resource group - whether data or resource, 
# can be easily referenced in other parts of the module.
###########################################################################################

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
