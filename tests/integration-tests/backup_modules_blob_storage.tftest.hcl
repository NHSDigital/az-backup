mock_provider "azurerm" {
  source = "./azurerm"
}

run "setup_tests" {
  module {
    source = "./setup"
  }
}

run "create_blob_storage_backup" {
  command = apply

  module {
    source = "../../infrastructure"
  }

  variables {
    vault_name     = run.setup_tests.vault_name
    vault_location = "uksouth"
  }

  assert {
    condition     = length(module.blob_storage_backup.backup_policy_id) > 0
    error_message = "Blob storage backup policy id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup.backup_policy_name == "bkpol-${var.vault_name}-blobstorage"
    error_message = "Blob storage backup policy name not as expected."
  }

  assert {
    condition     = module.blob_storage_backup.vault_id == azurerm_data_protection_backup_vault.backup_vault.id
    error_message = "Blob storage backup vault id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup.retention_period == "P7D"
    error_message = "Blob storage backup retention period not as expected."
  }

  assert {
    condition     = length(module.blob_storage_backup.azure_policy_definition_id) > 0
    error_message = "Blob storage backup azure policy definition id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup.azure_policy_definition_name == "policydef-${var.vault_name}-create-backup-instance-blob-storage"
    error_message = "Blob storage backup azure policy definition name not as expected."
  }

  assert {
    condition     = module.blob_storage_backup.azure_policy_definition_policy_type == "Custom"
    error_message = "Blob storage backup azure policy definition policy type not as expected."
  }

  assert {
    condition     = module.blob_storage_backup.azure_policy_definition_mode == "All"
    error_message = "Blob storage backup azure policy definition mode not as expected."
  }

  assert {
    condition     = length(module.blob_storage_backup.azure_policy_assignment_id) > 0
    error_message = "Blob storage backup azure policy assignment id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup.azure_policy_assignment_name == "policyass-${var.vault_name}-create-backup-instance-blob-storage"
    error_message = "Blob storage backup azure policy assignment name not as expected."
  }

  assert {
    condition     = module.blob_storage_backup.azure_policy_assignment_subscription_id == data.azurerm_subscription.current.id
    error_message = "Blob storage backup azure policy assignment subscription id not as expected."
  }

  assert {
    condition     = module.blob_storage_backup.azure_policy_assignment_location == var.vault_location
    error_message = "Blob storage backup azure policy assignment location not as expected."
  }

  assert {
    condition     = length(module.blob_storage_backup.azure_policy_assignment_identity[0].principal_id) > 0
    error_message = "Blob storage backup azure policy assignment identity not as expected."
  }

  assert {
    condition = module.blob_storage_backup.azure_policy_assignment_parameters == jsonencode({
      vaultName = {
        value = var.vault_name
      }
      backupInstanceName = {
        value = "bkinst-${var.vault_name}-blobstorage"
      }
      backupPolicyId = {
        value = module.blob_storage_backup.backup_policy_id
      }
    })
    error_message = "Blob storage backup azure policy assignment parameters not as expected."
  }
}