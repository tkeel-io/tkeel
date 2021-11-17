# PLUGINS-005-plugin-lifecycle

## Status
Proposed

## Context
Unified standardization of plug-in process status

## Decision

### Plug-in life cycle

The plug-in is divided into the following steps from installation to uninstallation

1. Download the helm charts and install to k8s
2. Register the plugin to the platform
3. Unregister the plugin in the platform
4. Uninstall helm charts from k8s

Tip: tKeel will focus on (paas manager \ tantent manager \ tantent user)

```
      developer         +        paas manager         +     tantent manager
                        |                             |
   +------------+       |       +-----------+         |      +----------+
   |            |       |       |           |         |      |          |
   | developing |       |       | published |         |      | disabled |
   |            |       |       |           |         |      |          |
   +----+-------+       |       +---+-------+         |      +---+------+
        |               |   install |                 |          |
        |  ^            |           v   ^             |          | ^
        |  |            |               | uninstall   |          | |
        |  |            |       +-------+---+         |          | |
release |  |            |       |           |         |   enable | |
        |  | upgrade    |       | installed |         |          | | disable
        |  |            |       |           |         |          | |
        |  |            |       +---+-------+         |          | |
        |  |            |  register |                 |          | |
        v  |            |           v  ^              |          v |
           |            |              | remove       |            |
   +-------+----+       |       +------+----+         |      +-----+----+
   |            |       |       |           |         |      |          |
   |  release   |       |       |registered |         |      | enabled  |
   |            |       |       |           |         |      |          |
   +------------+       +       +-----------+         +      +----------+

```

## Consequences
