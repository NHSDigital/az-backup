# Developer Guide

## Overview

The following guide is for developers working on the blueprint solution - not for developers that are consuming the blueprint.

## Environment Setup

The following are pre-requisites to working with the solution:

* An Azure subscription for development purposes
* An Azure identity which has been assigned the following roles at the subscription level:
    * Contributor (required to create resources)
    * Role Based Access Control Administrator (to assign roles to the backup vault managed identity) **with a condition that limits the roles which can be assigned to:**
        * Disk Backup Reader
        * Disk Snapshot Contributor
        * PostgreSQL Flexible Server Long Term Retention Backup Role
        * Storage Account Backup Contributor
        * Reader
* [Azure CLI installed](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli-windows?tabs=azure-cli)
* [Terraform installed](https://developer.hashicorp.com/terraform/install)
* [Go installed (to run the end-to-end tests)](https://go.dev/dl/)

Ensure all installed components have been added to the `%PATH%` - e.g. `az`, `terraform` and `go`.

## Getting Started

Take the following steps to get started in configuring and verifying the infrastructure for your development environment:

1. Setup environment variables

    Set the following environment variables in order to connect to Azure in the following steps:

    ```pwsh
    $env:ARM_TENANT_ID="<your-tenant-id>"
    $env:ARM_SUBSCRIPTION_ID="<your-subscription-id>"
    $env:ARM_CLIENT_ID="<your-client-id>"
    $env:ARM_CLIENT_SECRET="<your-client-secret>"
    ```

1. Create Backend

    A backend (e.g. storage account) is required in order to store the tfstate and work with Terraform.

    Run the following powershell script to create the backend with default settings: `./scripts/create-tf-backend.ps1`. This script will create a resource group called `rg-nhsbackup` containing a storage account called `satfstate<random-id>`.

    Make a note of the name of the storage account in the script output - it's generated with a random suffix, and you'll need it in the following steps to initialise the terraform.

1. Initialise Terraform

    Change the working directory to `./infrastructure`.

    Terraform can now be initialised by running the following command:

    ````pwsh
    terraform init -backend=true -backend-config="resource_group_name=rg-nhsbackup" -backend-config="storage_account_name=<storage-account-name>" -backend-config="container_name=tfstate" -backend-config="key=terraform.tfstate"
    ````

1. Prepare Terraform Variables

    You need to specify the mandatory terraform variables as a minimum, and may want to specify a number of the optional variables.

    You can specify the variables via the command line when executing `terraform apply`, or by preparing a tfvars file and specifying the path to that file.

    Here are examples of each approach:

    ```pwsh
    terraform apply -var resource_group_name=<resource-group-name> -var backup_vault_name=<backup-vault-name> var tags={"tagOne" = "tagOneValue"} -var blob_storage_backups={"backup1" = { "backup_name" = "myblob", "retention_period" = "P7D", "backup_intervals" = ["R/2024-01-01T00:00:00+00:00/P1D"], "storage_account_id" = "id" }}
    ```

    ```pwsh
    terraform apply -var-file="<your-var-file>.tfvars
    ```

1. Apply Terraform

    Apply the Terraform code to create the infrastructure.

    The `-auto-approve` flag is used to automatically approve the plan, you can remove this flag to review the plan before applying.

    ```pwsh
    terraform apply -auto-approve
    ```

    Now review the deployed infrastructure in the Azure portal. You will find the resources deployed to a resource group called `rg-nhsbackup-myvault` (unless you specified a different vault name in the tfvars).

    Should you want to, you can remove the infrastructure with the following command:

    ```pwsh
    terraform destroy -auto-approve
    ```

## Integration Tests

The test suite consists of a number Terraform HCL integration tests that use a mock azurerm provider.

[See this link for more information.](https://developer.hashicorp.com/terraform/language/tests)

> TIP! Consider adopting the classic red-green-refactor approach using the integration test framework when adding or modifying the terraform code.

Take the following steps to run the test suite:

1. Initialise Terraform

    Change the working directory to `./tests/integration-tests`.

    Terraform can now be initialised by running the following command:

    ````pwsh
    terraform init -backend=false
    ````

    > NOTE: There's no need to initialise a backend for the purposes of running the tests.

1. Run the tests

    Run the tests with the following command:

    ````pwsh
    terraform test
    ````

## End to End Tests

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

1. Setup environment variables

    The end-to-end test suite needs to login to Azure in order to execute the tests and therefore the following environment variables must be set.

    ```pwsh
    $env:ARM_TENANT_ID="<your-tenant-id>"
    $env:ARM_SUBSCRIPTION_ID="<your-subscription-id>"
    $env:ARM_CLIENT_ID="<your-client-id>"
    $env:ARM_CLIENT_SECRET="<your-client-secret>"
    $env:TF_STATE_RESOURCE_GROUP="rg-nhsbackup"
    $env:TF_STATE_STORAGE_ACCOUNT="<storage-account-name>"
    $env:TF_STATE_STORAGE_CONTAINER="tfstate"
    ```

    > For the storage account name, the TF state backend should have been created during the [getting started guide](#getting-started), at which point the storage account will have been created and the name generated.

1. Run the tests

    Run the tests with the following command:

    ````pwsh
    go test -v -timeout 10m
    ````

### Debugging

To debug the tests in vscode, add the following configuration to launch settings and run the configuration:

```json
{
    "configurations": [
        {
            "name": "Go Test",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}/tests/end-to-end-tests",
            "env": {
                "ARM_TENANT_ID": "<your-tenant-id>",
                "ARM_SUBSCRIPTION_ID": "<your-subscription-id>",
                "ARM_CLIENT_ID": "<your-client-id>",
                "ARM_CLIENT_SECRET": "<your-client-secret>",
                "TF_STATE_RESOURCE_GROUP": "rg-nhsbackup",
                "TF_STATE_STORAGE_ACCOUNT": "<storage-account-name>",
                "TF_STATE_STORAGE_CONTAINER": "tfstate"
            }
        }       
    ]
}
```

> For the storage account name, the TF state backend should have been created during the [getting started guide](#getting-started), at which point the storage account will have been created and the name generated.

## CI Pipeline

The CI pipeline builds and verifies the solution and runs a number of static code analysis steps on the code base.

Part of the build verification is end to end testing. This requires the pipeline to login to Azure and deploy an environment on which to execute the tests. In order for the pipeline to login to Azure the following GitHub actions secrets must be created:

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
