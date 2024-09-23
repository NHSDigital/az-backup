terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.114.0"
    }
    azapi = {
      source = "azure/azapi"
      version = "=1.15.0"
    }
  }

}

provider "azurerm" {
  features {}
}

provider "azapi" {}