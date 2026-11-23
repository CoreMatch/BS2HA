<div align="center">
  <h1>BS2HA</h1>
  <p><strong>从 BlessingSkin 迁移，让重要数据有据可查。</strong></p>
  <p><a href="./README.md">English</a></p>
</div>

BS2HA 是一款离线迁移工具，通过引导式 TUI 将 BlessingSkin 的 SQL 与纹理数据转换为 HRPAuth 核心和 Yggdrasil 产物。

它为计划受控迁移的管理员生成导入文件与迁移报告。

## 为什么选择 BS2HA

- **引导式转换。** 通过 TUI 协助管理员完成缺失项选择。
- **安全签名纹理。** 生成 RSA-4096 密钥与签名纹理产物。
- **保留可用密码。** 直接复制 bcrypt 哈希，无需重新计算。
- **同时生成两套导入。** 分别输出核心与 Yggdrasil SQL，并附迁移报告。
- **完全离线迁移。** 直接处理 SQL 转储与纹理文件。

## 进一步了解

- [迁移设计](docs/design.md) — 本地设计参考
- [Docker 镜像](https://hub.docker.com/r/limeight/ha-bs2ha) — 已发布的容器镜像
- [HA Wiki](https://wiki.mcnb.dev/) — 生态文档
- [HA-Contract](https://github.com/CoreMatch/HA-Contract) — 生态 API 契约

## 迁移限制

迁移不包含会话、令牌与 OAuth 客户端；非 bcrypt 密码需要重置。请先备份并测试——迁移绝非无损。

## 许可证与来源

[AGPL-3.0](LICENSE)。来源：[CoreMatch/BS2HA](https://github.com/CoreMatch/BS2HA)。
