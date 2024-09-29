# Azure Immutable Backup

## Introduction

This repository is a blueprint accelerator solution that supports teams in taking immutable backups in Azure. It's aim is to give developers tooling and templates that can be used to create, configure and manage immutable backups using Azure Backup Vault.

The following technologies are used:

* Azure
* Azure CLI
* Azure Pipelines
* Terraform
* Go (used for end-to-end testing)

### Outstanding Questions

* The design doesn't cater for the requirement to store the backup data in a separate account (or subscription in Azure lingo). We can however support GeoRedundant storage across regions - will this suffice? Otherwise we need to look at a solution for this problem.
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
* Entra ID
* Tfstate storage account
* Resources that need to be backed up

#### Architecture

The following diagram illustrates the high level architecture

![Azure Architecture](./docs/azure-architecture.drawio.svg)

1. The **backup vault** stores the backups of a variety of different Azure resources. A number of **backup policies** are registered on the vault which define the configuration for a backup such as the retention period and schedule. A number of **backup instances** are then registered with a policy applied that trigger the backups. The vault is configured as **immutable** and **locked** to enforce tamper proof backups. The **backup vault** resides in it's own isolated **resource group**.

1. **Backup instances** link the resources to be backed up and an associated **backup policy**, and one registered trigger the backup process. The resources directly supported are Azure Blob Storage, Managed Disks, PostgreSQL (single server and flexible server) and AKS instances, although other resources are supported indirectly through Azure Storage (see **point 8** for more details). **Backup instances** are automatically registered by **Azure Policy** by creating resources to be  backed up with the required tags - they are not manually registered (see **point 4** for more details).

1. The **backup vault** accesses resources to be backed up through a **System Assigned Managed Identity** - a secure way of enabling communication between defined resources without managing a secret/password. The identity is given read access to the resources to be backed up by **Azure Policy** at the point that the backup instance is registered.

1. **Azure Policy** is a feature that helps enforce rules and standards across an Azure tenant. In this case it is used to ensure **backup instances** are created when resources that require backup have a defined tag. **Azure Policy** will also be used to validate the **immutability** configuration of the backup vault, for example ensuring it is not set excessively resulting in a developers holiday photos being stored for 100'000 years.

1. **Backup administrators** are a group of identities that will have time limited read only access to the **backup vault** in order to access and restore backups as required. Assignment of the role will be secured by **PIM** - Privileged Identity Management, which requires a second identity to authorise the role assignment, which is then assigned on a time limited bases. The **backup administrators** will also be responsible for monitoring and auditing backup activity via **Azure Monitor** (see **point 7** for more details).

1. The solution requires a user account with elevated subscription contributor permissions that can create the backup resources (such as the backup **resource group**, **backup vault**, and **backup policies**). This identity will be implemented as a **federated credential** of an **app registration**, which is like a passport that lets you access different services without needing a separate password. This removes the need to manage a secret/password once configured. The identity also needs writer access to a dedicated **Storage Account** in order to read and write the **terraform** infrastructure state.

1. All backup telemetry will flow into **Azure Monitor** for monitoring and auditing purposes. This will provide access to data such as backup logs and metrics, and provide observability over the solution. Should the need arise, the telemetry could also be integrated into an external monitoring solution.

1. Some resources such as Azure SQL and Azure Key Vault are not directly supported by Azure **backup vault**, but can be incorporated via a supplementary process that backs up the data to Azure Blob Storage first. In the case of Azure SQL, a typical scenario could be an Azure Logic App that takes a backup of Azure SQL on a regular basis and stores the data in Azure Blob Storage.  It is the aspiration of this solution to provide guidance and tooling that teams can adopt to support these scenarios.

### Security Design

The following diagram illustrates the security design of the solution:

![Azure Architecture](./docs/security-design.drawio.svg)

See the following links for further details on some concepts relevant to the design:

* [Azure Multi-user Authorisation (MUA) and Resource Guard](https://learn.microsoft.com/en-us/azure/backup/multi-user-authorization-concept)
* [Backup Operator Role](https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles/storage#backup-operator)
* [Azure Privileged Identity Management (PIM)](https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management)

#### Actors

> NOTE: The roles listed below are not an exhaustive list, and are only those which are of relevance to the backup solution.

1. Tenant Admin

   The tenant admin, aka the "global administrator", is typically a restricted group of technical specialists and/or senior engineering staff. They have full control over the Azure tenant including all subscriptions and identities.

   The actor holds the following roles:

   * Tenant Owner
   * Tenant RBAC Administrator

   The following risks and mitigations should be considered:

   | Risks | Mitigations |
   |-|-|
   | Backup instance tampered with. | Use of PIM for temporary elevated privileges. |
   | Backup policy tampered with. | Use of MUA for restricted backup operations. |
   | Role based access tampered. | Dedicated admin accounts. |
   | No other account able to override a malicious actor. | |

1. Subscription Admin

   The subscription admin is typically a restricted group of team leads who are deploying their teams solutions to the subscription. They have full control over the subscription, including the backup vault and the backup resources.

   The actor holds the following roles:

   * Subscription Owner
   * Subscription RBAC Administrator

   The following risks and mitigations should be considered:

   | Risks | Mitigations |
   |-|-|
   | Backup instance tampered with.&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; | Use of PIM for temporary elevated privileges. |
   | Backup policy tampered with. | Use of MUA for restricted backup operations. |
   | Role based access tampered. | |

1. Deployment Service Principal

   The deployment service principal is an unattended credential used to deploy the solution from an automated process such as a pipeline or workflow. It has the permission to deploy resources (such as the backup vault) and assign the roles required for the solution to operate.

   The actor holds the following roles:

   * Subscription Contributor
   * Subscription RBAC Administrator limited to the roles required by the deployment  

   The following risks and mitigations should be considered:

   | Risks | Mitigations |
   |-|-|
   | Backup instance tampered with.&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; | Use of PIM for temporary elevated privileges. |
   | Backup policy tampered with. | Use of MUA for restricted backup operations. |
   | Role based access tampered. | Secret scanning in pipeline. |
   | Poor secret management. | Robust secret management procedures. |

1. Backup Admin

   The backup admin is typically a group of team support engineers and/or technical specialists. They have the permission to monitor backup telemetry, and restore backups in order to recover services.

   The actor holds the following roles:

   * Subscription Backup Operator

1. Security Admin

   The security admin is typically a group of cyber security specialists that are isolated from the other actors, by being in a different tenant or a highly restricted subscription. They have permissions to manage Resource Guard, which provide multi user authorisation to perform restricted operations on the backup vault, such as changing policies or stopping a backup instance.

   The actor holds the following roles:

   * Subscription Backup MUA Administrator

   | Risks | Mitigations |
   |-|-|
   | Elevated roles note revoked.&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; | Use of PIM for temporary elevated privileges. |
   |  | Robust and well documented processes. |

   **NOTE: MUA without PIM requires a manual revocation of elevated permissions.**

1. Backup Vault Managed Identity

   The backup vault managed identity is a "System Assigned" managed identity that performs backup vault operations. It is restricted to just the services defined at deployment, and cannot be compromised at runtime.

   The actor holds the following roles:

   * Backup Vault Resource Writer
   * Reader role on resources that require backup

### Pipelines

> TODO

## Repository Structure

The repository consists of the following directories:

* `./.github`
  
  Contains the GitHub workflows in `yaml` format.
  
  [See the YAML schema documentation for more details.](https://learn.microsoft.com/en-us/azure/devops/pipelines/yaml-schema/?view=azure-pipelines)

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

* `./tests`

  Contains the different types of tests used to verify the solution.

## Developer Guide

### Environment Setup

The following are pre-reqs to working with the solution:

* An Azure subscription
* An Azure identity which has been assigned the subscription Contributor role (required to create resources)
* [Azure CLI installed](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli-windows?tabs=azure-cli)
* [Terraform installed](https://developer.hashicorp.com/terraform/install)
* [Go installed (to run the end-to-end tests)](https://go.dev/dl/)

> Ensure all installed components have been added to the `%PATH%` - e.g. `az`, `terraform` and `go`.

### Getting Started

Take the following steps to get started in configuring and verifying the infrastructure for your development environment:

1. Setup environment variables

   Set the following environment variables in order to connect to Azure in the following steps:

   ```pwsh
   $env:ARM_TENANT_ID="<your-tenant-id>"
   $env:ARM_SUBSCRIPTION_ID="<your-subscription-id>"
   $env:ARM_CLIENT_ID="<your-client-id>"
   $env:ARM_CLIENT_SECRET="<your-client-secret>"
   ```

2. Create Backend

   A backend (e.g. storage account) is required in order to store the tfstate and work with Terraform.

   Run the following powershell script to create the backend with default settings: `./scripts/create-tf-backend.ps1`. This script will create a resource group called `rg-nhsbackup` containing a storage account called `satfstate<random-id>`.

   Make a note of the name of the storage account in the script output - it's generated with a random suffix, and you'll need it in the following steps to initialise the terraform.

3. Prepare Terraform Variables (Optional)

   If you want to override the Terraform variables, make a copy of `tfvars.template` and amend any default settings as required.

   In the next step add the following flag to the `terraform apply` command in order to use your variables:

   ```pwsh
   -var-file="<your-var-file>.tfvars
   ```

4. Initialise Terraform

   Change the working directory to `./infrastructure`.

   Terraform can now be initialised by running the following command:

   ````pwsh
   terraform init -backend=true -backend-config="resource_group_name=rg-nhsbackup" -backend-config="storage_account_name=<storage-account-name>" -backend-config="container_name=tfstate" -backend-config="key=terraform.tfstate"
   ````

5. Apply Terraform

   Apply the Terraform code to create the infrastructure.

   The `-auto-approve` flag is used to automatically approve the plan, you can remove this flag to review the plan before applying.

   ```pwsh
   terraform apply -auto-approve
   ```

   Now review the deployed infrastructure in the Azure portal. You will find a backup vault and some sample backup policies.

   The repo contains an `example` module which can be utilised to further extend the sample infrastructure with some resources and backup instances. To use this module for dev/test purposes, include the module in `main.tf` and run `terraform apply` again.

### Running the Tests

#### Integration Tests

The test suite consists of a number Terraform HCL integration tests that use a mock azurerm provider.

[See this link for more information.](https://developer.hashicorp.com/terraform/language/tests)

Take the following steps to run the test suite:

1. Initialise Terraform

   Change the working directory to `./tests/integration-tests`.

   Terraform can now be initialised by running the following command:

   ````pwsh
   terraform init -backend=false
   ````

   > NOTE: There's no need to initialise a backend for the purposes of running the tests.

2. Run the tests

   Run the tests with the following command:

   ````pwsh
   terraform test
   ````

#### End to End Tests

The end to end tests are written in go, and use the [terratest library](https://terratest.gruntwork.io/) and the [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go/tree/main).

The tests depend on a connection to Azure so it can create an environment that the tests can be executed against - the environment is torn down once the test run has completed.

See the following resources for docs and examples of terratest and the Azure SDK:

* [Terratest docs](https://terratest.gruntwork.io/docs/)
* [Terratest repository](https://github.com/gruntwork-io/terratest)
* [Terratest test examples](https://github.com/gruntwork-io/terratest/tree/master/test)
* [Azure SDK](https://github.com/Azure/azure-sdk-for-go/tree/main)
* [Azure SDK Data Protection Module](https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/resourcemanager/dataprotection/armdataprotection)

To run the tests, take the following steps:

1. Install go packages

   You only need to do this once when setting up your environment.

   Change the working directory to `./tests/end-to-end-tests`.

   Run the following command:

   ````pwsh
   go mod tidy
   ````

2. Setup environment variables

   The end-to-end test suite needs to login to Azure in order to execute the tests and therefore the following environment variables must be set.

   ```pwsh
   $env:ARM_TENANT_ID="<your-tenant-id>"
   $env:ARM_SUBSCRIPTION_ID="<your-subscription-id>"
   $env:ARM_CLIENT_ID="<your-client-id>"
   $env:ARM_CLIENT_SECRET="<your-client-secret>"
   $env:TF_STATE_RESOURCE_GROUP="rg-nhsbackup"
   $env:TF_STATE_STORAGE_ACCOUNT="<storage-account-name>"
   $env:TF_STATE_STORAGE_CONTAINER="terraform"
   ```

   > For the storage account name, the TF state backend should have been created during the [getting started guide](#getting-started), at which point the storage account will have been created and the name generated.

3. Run the tests

   Run the tests with the following command:

   ````pwsh
   go test -v -timeout 10m
   ````

##### Debugging

To debug the tests in vscode, add the following configuration to launch settings and run the configuration with the test file you want to debug open:

```json
{
    "configurations": [
        {
            "name": "Go Test",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${file}",
            "env": {
                "ARM_TENANT_ID": "<your-tenant-id>",
                "ARM_SUBSCRIPTION_ID": "<your-subscription-id>",
                "ARM_CLIENT_ID": "<your-client-id>",
                "ARM_CLIENT_SECRET": "<your-client-secret>",
                "TF_STATE_RESOURCE_GROUP": "rg-nhsbackup",
                "TF_STATE_STORAGE_ACCOUNT": "<storage-account-name>",
                "TF_STATE_STORAGE_CONTAINER": "terraform"
            }
        }       
    ]
}
```

> For the storage account name, the TF state backend should have been created during the [getting started guide](#getting-started), at which point the storage account will have been created and the name generated.

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

## CI Pipeline

The CI pipeline builds and verifies the solution and runs a number of static code analysis steps on the code base.

### End to End Testing

Part of the build verification is the end to end testing step. This requires the pipeline to login to Azure in order to deploy an environment on which to execute the tests.

A storage account must be provisioned with a container called `github-actions`, which is used by the CI pipeline to persist the terraform state.

In order for the CI pipeline to login to Azure and use the terraform state storage account, the following GitHub actions secrets must be created:

* `AZURE_TENANT_ID`

  The ID of an Azure tenant which can be used for the end to end test environment.

* `AZURE_SUBSCRIPTION_ID`

  The ID of an Azure subscription which can be used for the end to end test environment.

* `AZURE_CLIENT_ID`

  The client ID of an Azure service principal / app registration which can be used to authenticate with the end to end test environment.
  
  The app registration must have contributor permissions on the subscription in order to create resources.

* `AZURE_CLIENT_SECRET`

  The client secret of an Azure app registration which can be used to authenticate with the end to end test environment.

* `TF_STATE_RESOURCE_GROUP`

  The resource group which contains the TF state storage account.

* `TF_STATE_STORAGE_ACCOUNT`

  The storage account used for TF state.

* `TF_STATE_STORAGE_COMTAINER`

  The storage container used for TF state.

### Static Code Analysis

The following static code analysis checks are executed:

* [Terraform format](https://developer.hashicorp.com/terraform/cli/commands/fmt)
* [Terraform lint](https://github.com/terraform-linters/tflint)
* [Checkov scan](https://www.checkov.io/)
* [Gitleaks scan](https://github.com/gitleaks/gitleaks)
* [Trivy vulnerability scan](https://github.com/aquasecurity/trivy)
