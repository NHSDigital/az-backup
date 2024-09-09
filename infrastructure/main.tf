terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.114.0"
    }
  }

  backend "local" {
    path      = "terraform.tfstate"
    condition = var.use_local_backend
  }

  backend "azurerm" {
    condition = !var.use_local_backend
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "resource_group" {
  location = var.vault_location
  name     = "rg-nhsbackup-${var.vault_name}"
}

# Create the vault
###########################################################################

resource "azurerm_data_protection_backup_vault" "backup_vault" {
  name                = "bvault-${var.vault_name}"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = var.vault_location
  datastore_type      = "VaultStore"
  redundancy          = var.vault_redundancy
  soft_delete         = "Off"
  identity {
    type = "SystemAssigned"
  }
}


# Create some dummy resources
###########################################################################

module "dummy_storage_account_1" {
  source               = "./modules/dummy_resource/storage_account"
  location             = var.vault_location
  storage_account_name = "samystorage001"
  resource_group       = azurerm_resource_group.resource_group.name
}

module "dummy_storage_account_2" {
  source               = "./modules/dummy_resource/storage_account"
  location             = var.vault_location
  storage_account_name = "samystorage002"
  resource_group       = azurerm_resource_group.resource_group.name
}

module "dummy_managed_disk" {
  source         = "./modules/dummy_resource/managed_disk"
  location       = var.vault_location
  disk_name      = "disk-mydisk"
  resource_group = azurerm_resource_group.resource_group.name
}


# Create some backup policies
###########################################################################

module "blob_storage_policy" {
  source           = "./modules/backup_policy/blob_storage"
  policy_name      = "bkpol-${var.vault_name}-blobstorage"
  vault_id         = azurerm_data_protection_backup_vault.backup_vault.id
  retention_period = "P7D" # 7 days
  # NOTE - this blob policy has been configured for operational backup 
  # only, which continuously backs up data and does not need a schedule
}

module "managed_disk_policy" {
  source           = "./modules/backup_policy/managed_disk"
  policy_name      = "bkpol-${var.vault_name}-manageddisk"
  vault_id         = azurerm_data_protection_backup_vault.backup_vault.id
  retention_period = "P7D"                               # 7 days
  backup_intervals = ["R/2024-01-01T00:00:00+00:00/P1D"] # Once per day at 00:00
}


# Create some backup instances
###########################################################################

module "blob_storage_instance_1" {
  source             = "./modules/backup_instance/blob_storage"
  instance_name      = "bkinst-${var.vault_name}-mystorage001"
  vault_id           = azurerm_data_protection_backup_vault.backup_vault.id
  vault_location     = var.vault_location
  vault_principal_id = azurerm_data_protection_backup_vault.backup_vault.identity[0].principal_id
  policy_id          = module.blob_storage_policy.id
  storage_account_id = module.dummy_storage_account_1.id

  depends_on = [
    module.blob_storage_policy,
    module.dummy_storage_account_1
  ]
}

module "blob_storage_instance_2" {
  source             = "./modules/backup_instance/blob_storage"
  instance_name      = "bkinst-${var.vault_name}-mystorage002"
  vault_id           = azurerm_data_protection_backup_vault.backup_vault.id
  vault_location     = var.vault_location
  vault_principal_id = azurerm_data_protection_backup_vault.backup_vault.identity[0].principal_id
  policy_id          = module.blob_storage_policy.id
  storage_account_id = module.dummy_storage_account_2.id

  depends_on = [
    module.blob_storage_policy,
    module.dummy_storage_account_2
  ]
}

module "managed_disk_instance" {
  source                      = "./modules/backup_instance/managed_disk"
  instance_name               = "bkinst-${var.vault_name}-mydisk"
  vault_id                    = azurerm_data_protection_backup_vault.backup_vault.id
  vault_location              = var.vault_location
  vault_principal_id          = azurerm_data_protection_backup_vault.backup_vault.identity[0].principal_id
  policy_id                   = module.managed_disk_policy.id
  managed_disk_id             = module.dummy_managed_disk.id
  managed_disk_resource_group = azurerm_resource_group.resource_group

  depends_on = [
    module.managed_disk_policy,
    module.dummy_managed_disk
  ]
}
