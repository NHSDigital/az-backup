
<!-- markdownlint-disable MD024 -->

# Security

## Overview

The security of the solution relies on configuration at the tenant and subscription level which is outside of the control of this module.

The design proposed in this section acts as a best practice guide and it will be down to teams and programmes to implement the necessary controls and procedures.

## Design

The following diagram illustrates the security design of the solution:

![Security Design](assets/security-design.drawio.svg)

See the following links for further details on some concepts relevant to the design:

* [Azure Multi-user Authorisation (MUA) and Resource Guard](https://learn.microsoft.com/en-us/azure/backup/multi-user-authorization-concept)
* [Backup Operator Role](https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles/storage#backup-operator)
* [Azure Privileged Identity Management (PIM)](https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management)

### Actors

> NOTE: The roles listed below are not an exhaustive list, and are only those which are of relevance to the backup solution.

#### 1. Tenant Admin

The tenant admin, aka the "global administrator", is typically a restricted group of technical specialists and/or senior engineering staff. They have full control over the Azure tenant including all subscriptions and identities.

The actor holds the following roles:

* Tenant Owner
* Tenant RBAC Administrator

The following risks and mitigations should be considered:

| Risks | Mitigations |
|-|-|
| Backup instance tampered with. | Use of PIM for temporary elevated privileges. |
| Backup policy tampered with. | Use of MUA for restricted backup operations. |
| Role based access tampered. | Dedicated admin accounts. |
| No other account able to override a malicious actor. | |

#### 2. Subscription Admin

The subscription admin is typically a restricted group of team leads who are deploying their teams solutions to the subscription. They have full control over the subscription, including the backup vault and the backup resources.

The actor holds the following roles:

* Subscription Owner
* Subscription RBAC Administrator

The following risks and mitigations should be considered:

| Risks | Mitigations |
|-|-|
| Backup instance tampered with.&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; | Use of PIM for temporary elevated privileges. |
| Backup policy tampered with. | Use of MUA for restricted backup operations. |
| Role based access tampered. | |

#### 3. Deployment Service Principal

The deployment service principal is an unattended credential used to deploy the solution from an automated process such as a pipeline or workflow. It has the permission to deploy resources (such as the backup vault) and assign the roles required for the solution to operate.

The actor holds the following roles:

* Subscription Contributor
* Subscription RBAC Administrator limited to the roles required by the deployment  

The following risks and mitigations should be considered:

| Risks | Mitigations |
|-|-|
| Backup instance tampered with.&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; | Use of PIM for temporary elevated privileges. |
| Backup policy tampered with. | Use of MUA for restricted backup operations. |
| Role based access tampered. | Secret scanning in pipeline. |
| Poor secret management. | Robust secret management procedures. |

#### 4. Backup Admin

The backup admin is typically a group of team support engineers and/or technical specialists. They have the permission to monitor backup telemetry, and restore backups in order to recover services.

The actor holds the following roles:

* Subscription Backup Operator

#### 5. Backup Monitor

The backup monitor is typically a group of service staff. They have the permission to monitor backup telemetry in order to raise the alarm if any issues are found.

The actor holds the following roles:

* Monitoring Reader

#### 6. Security Admin

The security admin is typically a group of cyber security specialists that are isolated from the other actors, by being in a different tenant or a highly restricted subscription. They have permissions to manage Resource Guard, which provide multi user authorisation to perform restricted operations on the backup vault, such as changing policies or stopping a backup instance.

The actor holds the following roles:

* Subscription Backup MUA Administrator

| Risks | Mitigations |
|-|-|
| Elevated roles note revoked.&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; | Use of PIM for temporary elevated privileges. |
|  | Robust and well documented processes. |

**NOTE: MUA without PIM requires a manual revocation of elevated permissions.**

#### 7. Backup Vault Managed Identity

The backup vault managed identity is a "System Assigned" managed identity that performs backup vault operations. It is restricted to just the services defined at deployment, and cannot be compromised at runtime.

The actor holds the following roles:

* Backup Vault Resource Writer
* Reader role on resources that require backup
