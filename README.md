# Azure Immutable Backup

## Introduction

This repository is a blueprint solution for deploying immutable backups to Azure. It's aim is to give developers tooling and templates that can be used to create, configure and manage immutable backups using Azure Backup Vault and Azure Recovery Services Vault.

The following technologies are used:

* Azure
* Azure CLI
* Azure Pipelines
* Terraform

### Outstanding Questions

* The design currently caters for a scenario where a vault could be unlocked initially, and later locked. Do we want this?

## Design

The repository consists of:

* Terraform modules to create the infrastructure
* Azure Pipelines to manage the deployment

### Infrastructure

A solution which utilises the blueprint will consist of the following types of Azure resources

* Azure backup vault and backup policies/instances
* Azure policy definitions and assignments
* Azure monitor
* Resources that need to be backed up
* Tfstate storage account
* Entra ID

#### Architecture

The following diagram illustrates the high level architecture

![Azure Architecture](./docs/azure-architecture.drawio.svg)

1. Backup vault
1. Backup instances
1. Backup identity
1. Azure policy
1. Backup administrators
1. Service deployment
1. Azure monitor
1. Indirect backup resources

### Pipelines

> TODO

## Repository Structure

The repository consists of the following directories:

* `./.pipelines`
  
  Contains the Azure Pipelines in `yaml` format.
  
  [See the YAML schema documentation for more details.](https://learn.microsoft.com/en-us/azure/devops/pipelines/yaml-schema/?view=azure-pipelines)

* `./docs`

  Stores files and assets related to the documentation.

* `./infrastructure`

  Stores the infrastructure as code - a set of terraform scripts and modules.
  
  [See the Terraform AzureRM documentation for more details.](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)

  [Also see the backup instance for blob storage as an example of the all the components that make up a blob storage backup.](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/data_protection_backup_instance_blob_storage)

* `./scripts`

  Contains scripts that are used to create and maintain the environment.

## Developer Guide

### Environment Setup

The following are pre-reqs to working with the solution:

* An Azure subscription
* Azure CLI installed
* Terraform installed
* An Azure identity with the following roles:
  * Contributor role on the subscription (required to create resources)
  * RBAC Administrator role on the resources being backed up (required to assign roles on the resource to the backup vault managed identity)

[See the following link for further information.](https://learn.microsoft.com/en-us/azure/developer/terraform/get-started-windows-powershell)

### Getting Started

Take the following steps to get started in configuring and verify the infrastructure:

1. Login to Azure

   Use the Azure CLI to login to Azure by running the following command:

   ```pwsh
   az login
   ```

2. Create Backend

   A backend (e.g. storage account) is required in order to store the tfstate and work with Terraform.

   Run the following powershell script to create the backend with default settings: `./scripts/create-tf-backend.ps1`.

   Make a note of the name of the storage account - it's generated with a random suffix, and you'll need it in the following steps.

3. Create Access Key

   An access key must be created as an environment variable so Terraform can authenticate with Azure.

   Run the following commands to generate an access key and store it as an environment variable:

   ```pwsh
   $ACCOUNT_KEY=$(az storage account keys list --resource-group "rg-nhsbackup" --account-name "<storage-account-name>" --query '[0].value' -o tsv)

   $env:ARM_ACCESS_KEY=$ACCOUNT_KEY
   ```

4. Prepare Terraform Variables (Optional)

   If you want to override the Terraform variables, make a copy of `tfvars.template` and amend any default settings as required.

   In the next step add the following flag to the `terraform apply` command in order to use your variables:

   ```pwsh
   -var-file="<your-var-file>.tfvars
   ```

5. Initialise Terraform

   Change the working directory to `./infrastructure`.

   Terraform can now be initialised by running the following command:

   ````pwsh
   terraform init -backend=true -backend-config="resource_group_name=rg-nhsbackup" -backend-config="storage_account_name=<storage-account-name>" -backend-config="container_name=tfstate" -backend-config="key=terraform.tfstate"
   ````

6. Apply Terraform

   Apply the Terraform code to create the infrastructure.

   The `-auto-approve` flag is used to automatically approve the plan, you can remove this flag to review the plan before applying.

   ```pwsh
   terraform apply -auto-approve
   ```

   Now review the deployed infrastructure in the Azure portal. You will find a dummy scenario consisting of some storage accounts and a managed disk, with a backup vault, backup policies and some sample backup instances.

### Contributing

If you want to contribute to the project, raise a PR on GitHub.

We use pre-commit to run analysis and checks on the changes being committed. Take the following steps to ensure the pre-commit hook is installed and working:

1. Install git
    * Ensure the git `bin` directory has been added to %PATH%: `C:\Program Files\Git\bin`

1. Install Python
    * Ensure the python `bin` directory has been added to %PATH%

1. Install pre-commit
    * Open a terminal and navigate to the repository root directory
    * Install pre-commit with the following command: `pip install pre-commit`
    * Install pre-commit within the repository with the following command: `pre-commit install`
    * Run `pre-commit run --all-files` to check pre-commit is working

    > For full details [see this link](https://pre-commit.com/#installation)
