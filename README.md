# Azure Immutable Backup

![CI](https://github.com/nhsdigital/az-backup/actions/workflows/ci-pipeline.yaml/badge.svg)

## Introduction

This repository is a blueprint accelerator solution that supports teams in implementing immutable backups in Azure.

It's aim is to give developers tooling and templates that can be used to create, configure and manage immutable backups using Azure Backup Vault in a proven way that's consistent across the organisation.

See the following key docs for more information:

* [Design](./docs/design.md)
* [Usage](./docs/usage.md)
* [Developer Guide](./docs/developer-guide.md)

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

## Documentation

The documentation is built and served as a static site using dockerised [MkDocs](https://www.mkdocs.org/).

To build and run the docs locally, install Docker then run the following command from the root of the repository:

```pwsh
docker-compose -f ./docs/docker-compose.yml up
```

Once the container is running, navigate to [http://localhost:8000](http://localhost:8000) to access the docs.

## Contributing

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
