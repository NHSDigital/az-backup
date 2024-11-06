resource "azurerm_log_analytics_workspace" "log_analytics" {
  name                = "law-nhsbackup-example"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  sku                 = "PerGB2018"
  retention_in_days   = 30
}