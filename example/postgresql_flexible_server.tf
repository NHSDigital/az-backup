resource "azurerm_postgresql_flexible_server" "postgresql_flexible_server" {
  name                          = "psql-nhsbackup-example"
  resource_group_name           = azurerm_resource_group.resource_group.name
  location                      = azurerm_resource_group.resource_group.location
  version                       = "14"
  public_network_access_enabled = false
  administrator_login           = "supersecurelogin"
  administrator_password        = "supersecurepassword"
  zone                          = "1"
  storage_mb                    = 32768
  storage_tier                  = "P4"
  sku_name                      = "B_Standard_B1ms"
}