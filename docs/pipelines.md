# Pipelines

## CI Pipeline

The CI pipeline builds and verifies the solution and runs a number of static code analysis steps on the code base. Once successful, if the pipeline is running against the `main` branch a GitHub release will be published using [Semantic Release.](https://github.com/cycjimmy/semantic-release-action)

Part of the build verification is end to end testing which requires the pipeline to login to an Azure tenant and deploy an environment on which to execute the tests.

### Static Code Analysis

The following static code analysis checks are executed:

* [Terraform format](https://developer.hashicorp.com/terraform/cli/commands/fmt)
* [Terraform lint](https://github.com/terraform-linters/tflint)
* [Checkov scan](https://www.checkov.io/)
* [Gitleaks scan](https://github.com/gitleaks/gitleaks)
* [Trivy vulnerability scan](https://github.com/aquasecurity/trivy)

### Pipeline Secrets

 In order for the pipeline to login to Azure the following secrets must be created:

* `AZURE_TENANT_ID`
  
  The ID of an Azure tenant which can be used for the end to end test environment.

* `AZURE_SUBSCRIPTION_ID`
  
  The ID of an Azure subscription which can be used for the end to end test environment.

* `AZURE_CLIENT_ID`
  
  The client ID of an Azure service principal / app registration which can be used to authenticate with the end to end test environment.
  
  The app registration must have contributor permissions on the subscription in order to create resources, and RBAC admin as described in [Environment Setup](./developer-guide.md#environment-setup).

* `AZURE_CLIENT_SECRET`
  
  The client secret of an Azure app registration which can be used to authenticate with the end to end test environment.

* `TF_STATE_RESOURCE_GROUP`
  
  The resource group which contains the TF state storage account.

* `TF_STATE_STORAGE_ACCOUNT`
  
  The storage account used for TF state.

* `TF_STATE_STORAGE_CONTAINER`
  
  The storage container used for TF state.

For the release tag to be added to the repository the following secrets must be created:

* `RELEASE_TOKEN`
  
  A personal access token which allows the pipeline to commit a release tag to the repository. The PAT will expire periodically and must be maintained.

  The PAT should be a fine grained access token, restricted to the `az-backup` repository, with Read/Write for the "Contents" permission.
