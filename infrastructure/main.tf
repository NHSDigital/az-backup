terraform {
  required_version = "~> 1.9 , < 2.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.18.0"
    }
    azapi = {
      source  = "Azure/azapi"
      version = "1.15.0"
    }
  }

  backend "azurerm" {}
}

provider "azurerm" {
  features {}
}

provider "azapi" {}
