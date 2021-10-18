# KEEL-000-overview-and-roadmap

## Overview
**Keel** 插件是为平台提供统一的外部流量访问的 API 网关的能力，在平台中起到代理的角色。

对于外部流量，统一的流量入口可以更好的去管理流量的安全性和用户的访问也更加集中，让平台在外部更加整体化。

对于内部流量，插件与插件间的访问通过 **Keel** 进行反向代理来实现，让插件访问插件变成插件访问平台。对于开发者而言，仅关注与平台交互即可。

## Roadmap(Schedule)
1. 2021/09
* [内部流量](./KEEL-001-internal-flow.md)
* [外部流量](./KEEL-002-external-flow.md)
2. 2021/10
* [API 版本](./KEEL-003-api-version.md)