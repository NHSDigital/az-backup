resource "random_id" "storage_account" {
  byte_length = 2
}

resource "azurerm_storage_account" "storage_account_one" {
  name                     = "stnhsbackupexample1${random_id.storage_account.hex}"
  resource_group_name      = azurerm_resource_group.resource_group.name
  location                 = azurerm_resource_group.resource_group.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_container" "storage_account_one_container_one" {
  name                  = "container1"
  storage_account_name  = azurerm_storage_account.storage_account_one.name
  container_access_type = "private"
}

resource "azurerm_storage_container" "storage_account_one_container_two" {
  name                  = "container2"
  storage_account_name  = azurerm_storage_account.storage_account_one.name
  container_access_type = "private"
}

resource "azurerm_storage_account" "storage_account_two" {
  name                     = "stnhsbackupexample2${random_id.storage_account.hex}"
  resource_group_name      = azurerm_resource_group.resource_group.name
  location                 = azurerm_resource_group.resource_group.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_container" "storage_account_two_container_one" {
  name                  = "container1"
  storage_account_name  = azurerm_storage_account.storage_account_two.name
  container_access_type = "private"
}

resource "azurerm_storage_container" "storage_account_two_container_two" {
  name                  = "container2"
  storage_account_name  = azurerm_storage_account.storage_account_two.name
  container_access_type = "private"
}