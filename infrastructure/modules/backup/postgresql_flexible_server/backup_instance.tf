resource "azurerm_role_assignment" "role_assignment_reader" {
  count                = var.assign_resource_group_level_roles == true ? 1 : 0
  scope                = var.server_resource_group_id
  role_definition_name = "Reader"
  principal_id         = var.vault.identity[0].principal_id
}

resource "azurerm_role_assignment" "role_assignment_long_term_retention_backup_role" {
  scope                = var.server_id
  role_definition_name = "PostgreSQL Flexible Server Long Term Retention Backup Role"
  principal_id         = var.vault.identity[0].principal_id
}

resource "azurerm_data_protection_backup_instance_postgresql_flexible_server" "backup_instance" {
  name             = "bkinst-pgflex-${var.backup_name}"
  vault_id         = var.vault.id
  location         = var.vault.location
  server_id        = var.server_id
  backup_policy_id = azurerm_data_protection_backup_policy_postgresql_flexible_server.backup_policy.id

  depends_on = [
    azurerm_role_assignment.role_assignment_reader,
    azurerm_role_assignment.role_assignment_long_term_retention_backup_role
  ]
}
