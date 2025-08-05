resource "azurerm_role_assignment" "role_assignment_snapshot_contributor" {
  count                = var.assign_resource_group_level_roles == true ? 1 : 0
  scope                = var.managed_disk_resource_group.id
  role_definition_name = "Disk Snapshot Contributor"
  principal_id         = var.vault.identity[0].principal_id
  principal_type       = "ServicePrincipal"
}

resource "azurerm_role_assignment" "role_assignment_backup_reader" {
  scope                = var.managed_disk_id
  role_definition_name = "Disk Backup Reader"
  principal_id         = var.vault.identity[0].principal_id
  principal_type       = "ServicePrincipal"
}

resource "azurerm_data_protection_backup_instance_disk" "backup_instance" {
  name                         = "bkinst-disk-${var.backup_name}"
  vault_id                     = var.vault.id
  location                     = var.vault.location
  disk_id                      = var.managed_disk_id
  snapshot_resource_group_name = var.managed_disk_resource_group.name
  backup_policy_id             = azurerm_data_protection_backup_policy_disk.backup_policy.id

  depends_on = [
    azurerm_role_assignment.role_assignment_snapshot_contributor,
    azurerm_role_assignment.role_assignment_backup_reader
  ]
}
