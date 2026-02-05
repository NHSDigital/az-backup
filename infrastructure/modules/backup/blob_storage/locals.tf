locals {
  # Render names using templates
  backup_policy_name = replace(
    replace(var.backup_policy_naming_template, "{resource_abbreviation}", "bkpol"),
    "{backup_name}",
    var.backup_name
  )

  backup_instance_name = replace(
    replace(var.backup_instance_naming_template, "{resource_abbreviation}", "bkinst"),
    "{backup_name}",
    var.backup_name
  )
}
