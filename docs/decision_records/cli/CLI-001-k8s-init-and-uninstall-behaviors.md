# CLI-002: Kubernetes mode Init and Uninstall behaviours

## Status
Accepted

## Context
The behavior of `init` and `uninstall` on Kubernetes mode for. 

## Decisions

* Calling `tkeel init` will
  * Check `dapr` Runtime. If the Namespace of Kubernetes is incorrect, a correct prompt will be given.
  * Download and install the core plugins of Tkeel, including Plugins(Plugin-Manager), Auth, Keel(API-Gateway).
  * Register core plugin, including Plugins(Plugin-Manager), Auth, Keel(API-Gateway).
* Calling `tkeel uninstall` will
  * Deregister core plugins, including Plugins, Auth, API-Gateway (Keel).
  * Remove the core plugins of TKeel, including Plugins, Auth, API-Gateway (Keel).

