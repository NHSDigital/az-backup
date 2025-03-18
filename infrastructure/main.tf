terraform {
  required_version = "~> 1.9 , < 2.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 4.18.0, < 5.0"
    }
  }

  backend "azurerm" {}
}
