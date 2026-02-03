locals {
  backup_instance_name = replace(replace(var.backup_instance_naming_template, "{resource_abbreviation}", "bkinst"), "{backup_name}", var.backup_name)
}
