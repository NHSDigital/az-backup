resource "azurerm_managed_disk" "managed_disk" {
  name                 = var.disk_name
  resource_group_name  = var.resource_group
  location             = var.location
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"
}
