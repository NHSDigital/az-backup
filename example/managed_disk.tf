resource "azurerm_managed_disk" "managed_disk" {
  name                 = "disk-nhsbackup-example"
  resource_group_name  = azurerm_resource_group.resource_group.name
  location             = azurerm_resource_group.resource_group.location
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"
}