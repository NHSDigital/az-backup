locals {

  resource_type = "disk"

  # Render names using templates
  backup_policy_name = replace(
    replace(
      replace(var.backup_policy_naming_template, "{resource_abbreviation}", "bkpol"),
      "{resource_type}", local.resource_type
    ),
    "{backup_name}", var.backup_name
  )

  backup_instance_name = replace(
    replace(
      replace(var.backup_instance_naming_template, "{resource_abbreviation}", "bkinst"),
      "{resource_type}", local.resource_type
    ),
    "{backup_name}", var.backup_name
  )

}
