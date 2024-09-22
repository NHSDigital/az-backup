resource "azurerm_policy_definition" "create_backup_instance" {
  name         = "policydef-${var.vault_name}-backup-blob-storage"
  policy_type  = "Custom"
  mode         = "All"
  display_name = "[AZ-BACKUP] Configure backup on blob storage accounts with a given tag"

  metadata = <<METADATA
  {
    "category": "Az-Backup"
  }
  METADATA

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
            "/providers/Microsoft.Authorization/roleDefinitions/00c29273-979b-4161-815c-10b084fb9324",
            "/providers/Microsoft.Authorization/roleDefinitions/f58310d9-a9f6-439a-9e8d-f62e7b41a168"
        ],
        "deployment": {
          "properties": {
            "mode": "incremental",
            "parameters": {
              "backupVaultId": {
                "value": "[parameters('backupVaultId')]"
              },
              "backupPolicyId": {
                "value": "[parameters('backupPolicyId')]"
              },
              "backupInstanceName": {
                "value": "[parameters('backupInstanceName')]"
              },
              "storageAccountId": {
                "value": "[field('id')]"
              }
            },
            "template": {
              "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
              "contentVersion": "1.0.0.0",
              "resources": [
                {
                  "type": "Microsoft.Authorization/roleAssignments",
                  "apiVersion": "2020-04-01",
                  "name": "[guid(parameters('storageAccountId'), 'StorageAccountBackupContributor')]",
                  "properties": {
                    "roleDefinitionName": "Storage Account Backup Contributor",
                    "principalId": "[reference(parameters('backupVaultId')).identity.principalId]",
                    "scope": "[parameters('storageAccountId')]"
                  }
                },  
                {
                  "type": "Microsoft.DataProtection/backupVaults/backupInstances",
                  "apiVersion": "2023-01-01",
                  "name": "[parameters('backupInstanceName')]",
                  "dependsOn": [
                    "[resourceId('Microsoft.Authorization/roleAssignments', guid(parameters('storageAccountId'), 'StorageAccountBackupContributor'))]"
                  ],
                  "properties": {
                    "dataSourceInfo": {
                      "resourceId": "[parameters('storageAccountId')]",
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
    "backupVaultId": {
      "type": "String",
      "metadata": {
        "description": "Resource ID of the backup vault"
      }
    },
    "backupPolicyId": {
      "type": "String",
      "metadata": {
        "description": "Resource ID of the backup policy to assign to the backup instance"
      }
    },
    "backupInstanceName": {
      "type": "String",
      "metadata": {
        "description": "Name of the backup instance to create"
      }
    }
  }
  PARAMETERS
}
