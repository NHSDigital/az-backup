run "create_blob_storage_policy" {
  command = apply

  module {
    source = "../infrastructure"
  }

  variables {
    vault_name = "testvault"
  }

  # Check that the id is as expected
  assert {
    condition     = length(module.blob_storage_policy.id) > 0
    error_message = "Blob storage policy id not as expected."
  }
}

run "create_managed_disk_policy" {
  command = apply

  module {
    source = "../infrastructure"
  }

  variables {
    vault_name = "testvault"
  }

  # Check that the id is as expected
  assert {
    condition     = length(module.managed_disk_policy.id) > 0
    error_message = "Managed disk policy id not as expected."
  }
}