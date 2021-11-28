# Architecture Decision Records

Architecture Decision Records (ADRs or simply decision records) are a collection of records for "architecturally significant" decisions. A decision record is a short markdown file in a specific light-weight format.

This folder contains all the decisions we have recorded in TKeel, including TKeel platform, TKeel CLI as well as TKeel SDKs in different languages.

## TKeel decision record organization and index

All decisions are categorized in the following folders:

* **Architecture** - Decisions on general architecture, code structure, coding conventions and common practices.

  
* **OPENAPI** - Decisions on TKeel platform OPENAPI designs.


* **CLI** - Decisions on TKeel CLI architecture and behaviors.

  - [CLI-001: Kubernetes mode Init and Uninstall behaviours](./cli/CLI-001-k8s-init-and-uninstall-behaviors.md)

* **SDKs** - Decisions on TKeel SDKs.


* **Engineering** - Decisions on Engineering practices, including CI/CD, testing and releases.

  - [ENG-001: Image Tagging](./engineering/ENG-001-tagging.md)

* **Plugin-Auth**(Abandoned) - Decisions on the core plugin Auth of TKeel plartform.

  - [AUTH-000: roadmap and glossary](./plugin-auth/AUTH-000-overview-and-roadmap.md)
  - [AUTH-001: tkeel certification management](./plugin-auth/AUTH-001-tkeel-certification-management.md)
  - [AUTH-002: tenant certification](./plugin-auth/AUTH-002-tenant-certification.md)
  - [AUTH-003: oauth2 design](./plugin-auth/AUTH-003-oauth2-design.md)

* **Plugin-Keel**(Abandoned) - Decisions on the core plugin keel of TKeel plartform.

  - [KEEL-000: Overview and roadmap](./plugin-keel/KEEL-000-overview-and-roadmap.md)
  - [KEEL-001: Internal flow](./plugin-keel/KEEL-001-internal-flow.md)
  - [KEEL-002: External flow](./plugin-keel/KEEL-002-external-flow.md)
  - [KEEL-003: api version](./plugin-keel/KEEL-003-api-version.md)

* **Plugin-Plugins**(Abandoned) - Decisions on the core plugin Plugins of TKeel plartform.

  - [PLUGINS-000: Overview and roadmap](./plugin-plugins/PLUGINS-000-overview-and-roadmap.md)
  - [PLUGINS-001: plugin data storage design](./plugin-plugins/PLUGINS-001-plugin-data-storage-design.md)
  - [PLUGINS-002: plugin management api design](./plugin-plugins/PLUGINS-002-plugin-management-api-design.md)
  - [PLUGINS-003: tenant management api design](./plugin-plugins/PLUGINS-003-tenant-management-api-design.md)
  - [PLUGINS-004: platform dependent version check](./plugin-plugins/PLUGINS-004-platform-dependent-version-check.md)


## Creating new decision records

A new decision record should be a _.md_ file named as 
```
<category prefix>-<sequence number in category>-<descriptive title>.md
```
|Category|Prefix|Annotation|
|----|----|----|
|Architecture|ARC|-|
|OPENAPI|OPENAPI|-|
|PLUGIN-AUTH|AUTH|ABANDONED|
|PLUGIN-KEEL|KEEL|ABANDONED|
|PLUGIN-PLUGINS|PLUGINS|ABANDONED|
|COMPONENT-RUDDER|RUDDER|-|
|COMPONENT-KEEL|KEEL|-|
|CLI|CLI|-|
|SDKs|SDK|-|
|Engineering|ENG|-|

A decision record should contain the following fields:

* **Status** - can be "proposed", "accepted", "implemented", or "rejected".
* **Context** - the context of the design discussion.
* **Decision** - Description of the decision.
* **Consequences** - what impacts this decision may create.
* **Implementation** - when a decision is implemented, the corresponding doc should be updated with the following information (when applicable):
  * Release version
  * Associated test cases
