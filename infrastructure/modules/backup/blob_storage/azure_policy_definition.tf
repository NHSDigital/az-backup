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
        "type": "Microsoft.Storage/storageAccounts/blobServices",
        "name": "default",
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
              "parameters": {
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
                },
                "storageAccountId": {
                  "type": "String",
                  "metadata": {
                    "description": "ID of the storage account to backup"
                  }
                }
              },
              "variables": {
                "storageAccountName": "[first(skip(split(parameters('storageAccountId'), '/'), 8))]",
                "dataSourceType": "Microsoft.Storage/storageAccounts/blobServices",
                "resourceType": "Microsoft.Storage/storageAccounts",
                "backupPolicyName": "[first(skip(split(parameters('backupPolicyId'), '/'), 10))]",
                "vaultName": "[first(skip(split(parameters('backupPolicyId'), '/'), 8))]",
                "vaultResourceGroup": "[first(skip(split(parameters('backupPolicyId'), '/'), 4))]",
                "vaultSubscriptionId": "[first(skip(split(parameters('backupPolicyId'), '/'), 2))]"
              },
              "resources": [ 
                {
                  "type": "Microsoft.Resources/deployments",
                  "apiVersion": "2021-04-01",
                  "resourceGroup": "[variables('vaultResourceGroup')]",
                  "subscriptionId": "[variables('vaultSubscriptionId')]",
                  "name": "[concat('DeployProtection-',uniqueString(variables('storageAccountName')))]",
                  "properties": {
                    "mode": "Incremental",
                    "template": {
                      "$schema": "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json#",
                      "contentVersion": "1.0.0.0",
                      "parameters": {},
                      "resources": [
                        {
                          "type": "Microsoft.Authorization/roleAssignments",
                          "apiVersion": "2022-04-01",
                          "name": "[guid(parameters('storageAccountId'), 'StorageAccountBackupContributor')]",
                          "properties": {
                            "roleDefinitionName": "Storage Account Backup Contributor",
                            "principalId": "[reference(parameters('backupVaultId')).identity.principalId]",
                            "scope": "[parameters('storageAccountId')]"
                          }
                        }, 
                        {
                          "type": "Microsoft.DataProtection/backupvaults/backupInstances",
                          "apiVersion": "2021-01-01",
                          "name": "[concat(variables('vaultName'), '/', variables('storageAccountName'))]",
                          "properties": {
                            "objectType": "BackupInstance",
                            "dataSourceInfo": {
                              "objectType": "Datasource",
                              "resourceID": "[parameters('storageAccountId')]",
                              "resourceName": "[variables('storageAccountName')]",
                              "resourceType": "[variables('resourceType')]",
                              "resourceUri": "[parameters('storageAccountId')]",
                              "resourceLocation": "[parameters('location')]",
                              "datasourceType": "[variables('dataSourceType')]"
                            },
                            "policyInfo": {
                              "policyId": "[parameters('backupPolicyId')]",
                              "name": "[variables('backupPolicyName')]"
                            }
                          }
                        }
                      ]
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
