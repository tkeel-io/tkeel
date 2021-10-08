# Architecture Decision Records

Architecture Decision Records (ADRs or simply decision records) are a collection of records for "architecturally significant" decisions. A decision record is a short markdown file in a specific light-weight format.

This folder contains all the decisions we have recorded in TKeel, including TKeel platform, TKeel CLI as well as TKeel SDKs in different languages.

## TKeel decision record organization and index

All decisions are categorized in the following folders:

* **Architecture** - Decisions on general architecture, code structure, coding conventions and common practices.

  
* **API** - Decisions on TKeel platform API designs.


* **CLI** - Decisions on TKeel CLI architecture and behaviors.

  - [CLI-001: Kubernetes mode Init and Uninstall behaviours](./cli/CLI-001-k8s-init-and-uninstall-behaviors.md)

* **SDKs** - Decisions on TKeel SDKs.


* **Engineering** - Decisions on Engineering practices, including CI/CD, testing and releases.

  - [ENG-001: Image Tagging](./engineering/ENG-001-tagging.md)

## Creating new decision records

A new decision record should be a _.md_ file named as 
```
<category prefix>-<sequence number in category>-<descriptive title>.md
```
|Category|Prefix|
|----|----|
|Architecture|ARC|
|API|API|
|CLI|CLI|
|SDKs|SDK|
|Engineering|ENG|

A decision record should contain the following fields:

* **Status** - can be "proposed", "accepted", "implemented", or "rejected".
* **Context** - the context of the design discussion.
* **Decision** - Description of the decision.
* **Consequences** - what impacts this decision may create.
* **Implementation** - when a decision is implemented, the corresponding doc should be updated with the following information (when applicable):
  * Release version
  * Associated test cases
