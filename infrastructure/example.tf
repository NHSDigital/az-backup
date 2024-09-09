# Create some example resources
###########################################################################

module "example_storage_account_1" {
  source               = "./modules/example_resource/storage_account"
  location             = var.vault_location
  storage_account_name = "samystorage001"
  resource_group       = azurerm_resource_group.resource_group.name
}

module "example_storage_account_2" {
  source               = "./modules/example_resource/storage_account"
  location             = var.vault_location
  storage_account_name = "samystorage002"
  resource_group       = azurerm_resource_group.resource_group.name
}

module "example_managed_disk" {
  source         = "./modules/example_resource/managed_disk"
  location       = var.vault_location
  disk_name      = "disk-mydisk"
  resource_group = azurerm_resource_group.resource_group.name
}


# Create some backup instances
###########################################################################

# NOTE - in future the backup instances will be created by Azure Policy 
# and will not need to be explictly defined in Terraform.

module "blob_storage_instance_1" {
  source             = "./modules/backup_instance/blob_storage"
  instance_name      = "bkinst-${var.vault_name}-mystorage001"
  vault_id           = azurerm_data_protection_backup_vault.backup_vault.id
  vault_location     = var.vault_location
  vault_principal_id = azurerm_data_protection_backup_vault.backup_vault.identity[0].principal_id
  policy_id          = module.blob_storage_policy.id
  storage_account_id = module.example_storage_account_1.id

  depends_on = [
    module.blob_storage_policy,
    module.example_storage_account_1
  ]
}

module "blob_storage_instance_2" {
  source             = "./modules/backup_instance/blob_storage"
  instance_name      = "bkinst-${var.vault_name}-mystorage002"
  vault_id           = azurerm_data_protection_backup_vault.backup_vault.id
  vault_location     = var.vault_location
  vault_principal_id = azurerm_data_protection_backup_vault.backup_vault.identity[0].principal_id
  policy_id          = module.blob_storage_policy.id
  storage_account_id = module.example_storage_account_2.id

  depends_on = [
    module.blob_storage_policy,
    module.example_storage_account_2
  ]
}

module "managed_disk_instance" {
  source                      = "./modules/backup_instance/managed_disk"
  instance_name               = "bkinst-${var.vault_name}-mydisk"
  vault_id                    = azurerm_data_protection_backup_vault.backup_vault.id
  vault_location              = var.vault_location
  vault_principal_id          = azurerm_data_protection_backup_vault.backup_vault.identity[0].principal_id
  policy_id                   = module.managed_disk_policy.id
  managed_disk_id             = module.example_managed_disk.id
  managed_disk_resource_group = azurerm_resource_group.resource_group

  depends_on = [
    module.managed_disk_policy,
    module.example_managed_disk
  ]
}
