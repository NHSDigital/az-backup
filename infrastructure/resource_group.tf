resource "azurerm_resource_group" "resource_group" {
  location = var.vault_location
  name     = "rg-nhsbackup-${var.vault_name}"
}
