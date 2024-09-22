resource "azurerm_policy_definition" "create_backup_instance" {
  name         = "policydef-${var.vault_name}-create-backup-instance-blob-storage"
  policy_type  = "Custom"
  mode         = "All"
  display_name = "[AZ-BACKUP] Configure backup on blob storage accounts with a given tag"

  policy_rule = <<POLICY_RULE
  {
    "if": {
      "allOf": [
        {
          "field": "type",
          "equals": "Microsoft.Storage/storageAccounts"
        },
        {
          "field": "tags['nhsbackup']",
          "equals": "enabled"
        }
      ]
    },
    "then": {
      "effect": "deployIfNotExists",
      "details": {
        "type": "Microsoft.DataProtection/backupVaults/backupInstances",
        "existenceCondition": {
          "allOf": [
            {
              "field": "Microsoft.DataProtection/backupVaults/backupInstances/dataSourceInfo.resourceID",
              "equals": "[field('id')]"
            },
            {
              "field": "Microsoft.DataProtection/backupVaults/backupInstances/policyInfo.policyId",
              "equals": "[parameters('backupPolicyId')]"
            }
          ]
        },
        "roleDefinitionIds": [
          "/providers/Microsoft.Authorization/roleDefinitions/4a333f42-bcae-4445-8538-3ec9ef8ad1f6"
        ],
        "deployment": {
          "properties": {
            "mode": "incremental",
            "parameters": {
              "vaultName": {
                "value": "[parameters('vaultName')]"
              },
              "backupPolicyId": {
                "value": "[parameters('backupPolicyId')]"
              },
              "backupInstanceName": {
                "value": "[parameters('backupInstanceName')]"
              }
            },
            "template": {
              "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
              "contentVersion": "1.0.0.0",
              "resources": [
                {
                  "type": "Microsoft.DataProtection/backupVaults/backupInstances",
                  "apiVersion": "2023-01-01",
                  "name": "[concat(parameters('vaultName'), '/', parameters('backupInstanceName'))]",
                  "properties": {
                    "dataSourceInfo": {
                      "resourceId": "[field('id')]",
                      "resourceType": "Microsoft.Storage/storageAccounts",
                      "dataSourceType": "AzureBlob"
                    },
                    "policyInfo": {
                      "policyId": "[parameters('backupPolicyId')]"
                    }
                  }
                }
              ]
            }
          }
        }
      }
    }
  }
  POLICY_RULE

  parameters = <<PARAMETERS
  {
    "vaultName": {
      "type": "string",
      "metadata": {
        "description": "Name of the backup vault"
      }
    },
    "backupPolicyId": {
      "type": "string",
      "metadata": {
        "description": "Resource ID of the backup policy to assign to the backup instance"
      }
    },
    "backupInstanceName": {
      "type": "string",
      "metadata": {
        "description": "Name of the backup instance to create"
      }
    }
  }
  PARAMETERS
}
