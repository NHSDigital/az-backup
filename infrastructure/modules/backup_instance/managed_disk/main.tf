resource "azurerm_role_assignment" "role_assignment_snapshot_contributor" {
  scope                = var.managed_disk_resource_group.id
  role_definition_name = "Disk Snapshot Contributor"
  principal_id         = var.vault_principal_id
}

resource "azurerm_role_assignment" "role_assignment_backup_reader" {
  scope                = var.managed_disk_id
  role_definition_name = "Disk Backup Reader"
  principal_id         = var.vault_principal_id
}

resource "azurerm_data_protection_backup_instance_disk" "backup_instance" {
  name                         = var.instance_name
  vault_id                     = var.vault_id
  location                     = var.vault_location
  disk_id                      = var.managed_disk_id
  snapshot_resource_group_name = var.managed_disk_resource_group.name
  backup_policy_id             = var.policy_id

  depends_on = [
    azurerm_role_assignment.role_assignment_snapshot_contributor,
    azurerm_role_assignment.role_assignment_backup_reader
  ]
}
