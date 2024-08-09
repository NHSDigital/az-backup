resource "azurerm_role_assignment" "extension_and_storage_account_permission" {
  scope                = var.storage_account_id
  role_definition_name = "Storage Account Contributor"
  principal_id         = var.cluster_extension_principal_id
}

resource "azurerm_role_assignment" "vault_msi_read_on_cluster" {
  scope                = var.cluster_id
  role_definition_name = "Reader"
  principal_id         = var.vault_principal_id
}

resource "azurerm_role_assignment" "vault_msi_read_on_snapshot_rg" {
  scope                = var.snapshot_resource_group_id
  role_definition_name = "Reader"
  principal_id         = var.vault_principal_id
}

resource "azurerm_role_assignment" "vault_msi_snapshot_contributor_on_snapshot_rg" {
  scope                = var.snapshot_resource_group_id
  role_definition_name = "Disk Snapshot Contributor"
  principal_id         = var.vault_principal_id
}

resource "azurerm_role_assignment" "vault_data_operator_on_snapshot_rg" {
  scope                = var.storage_account_id
  role_definition_name = "Data Operator for Managed Disks"
  principal_id         = var.vault_principal_id
}

resource "azurerm_role_assignment" "vault_data_contributor_on_storage" {
  scope                = var.snapshot_resource_group_id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = var.vault_principal_id
}

resource "azurerm_role_assignment" "cluster_msi_contributor_on_snapshot_rg" {
  scope                = azurerm_resource_group.snap.id
  role_definition_name = "Contributor"
  principal_id         = var.cluster_extension_principal_id
}

resource "azurerm_data_protection_backup_instance_kubernetes_cluster" "example" {
  name                         = var.instance_name
  location                     = var.vault_location
  vault_id                     = var.vault_id
  kubernetes_cluster_id        = var.cluster_id
  snapshot_resource_group_name = var.snapshot_resource_group_name
  backup_policy_id             = var.policy_id

  backup_datasource_parameters {
    excluded_namespaces              = ["test-excluded-namespaces"]
    excluded_resource_types          = ["exvolumesnapshotcontents.snapshot.storage.k8s.io"]
    cluster_scoped_resources_enabled = true
    included_namespaces              = ["test-included-namespaces"]
    included_resource_types          = ["involumesnapshotcontents.snapshot.storage.k8s.io"]
    label_selectors                  = ["kubernetes.io/metadata.name:test"]
    volume_snapshot_enabled          = true
  }

  depends_on = [
    azurerm_role_assignment.extension_and_storage_account_permission,
    azurerm_role_assignment.vault_msi_read_on_cluster,
    azurerm_role_assignment.vault_msi_read_on_snapshot_rg,
    azurerm_role_assignment.cluster_msi_contributor_on_snapshot_rg,
    azurerm_role_assignment.vault_msi_snapshot_contributor_on_snapshot_rg,
    azurerm_role_assignment.vault_data_operator_on_snapshot_rg,
    azurerm_role_assignment.vault_data_contributor_on_storage,
  ]
}
