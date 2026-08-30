# BS2HA 迁移工具设计说明

## 目标

做一个离线迁移工具：

- 输入 BlessingSkin 的数据库 SQL dump
- 输入 BlessingSkin 的 `textures/` 目录
- 输出可导入 HRPAuth 的 SQL 文件
- 输出迁移报告

目标是让迁移后的 HA 可以独立运行，不依赖旧 BS 继续在线。

## 一期范围

一期只处理以下目标表：

- `users`
- `profiles`
- `profile_properties`

不迁移运行态数据：

- `sessions`
- `tokens`
- `oauth2_*`

## 输入物

一期支持以下参数输入：

1. `--bs-sql <path>`：BlessingSkin SQL dump（默认：`./bs.sql`）
2. `--bs-textures <path>`：BlessingSkin 材质目录（默认：`./storage/textures`，若不存在则跳过材质迁移）
3. `--ha-callback-url <url>`：HRPAuth 回调基地址，用于生成 `textures.url`
4. `--ha-private-key <path>`：HRPAuth 私钥文件路径
5. `--ha-private-key-content <content>`：HRPAuth 私钥 PEM 内容（支持直接粘贴）
6. `--ha-baseline-sql <path>`：HRPAuth 数据库模板文件（默认：`./ha-baseline.sql`，由程序内嵌并自动释放）
7. `--out-sql <path>`：输出 SQL 文件路径
8. `--out-report <path>`：输出报告文件路径

## 交互设计 (TUI)

工具在缺失必要参数时会自动启动交互式界面 (TUI)：

- **高对比度显示**：适配各种终端配色方案。
- **自动路径探测**：自动扫描当前目录下的 `bs.sql`、材质目录和密钥文件，并优先提示用户确认。
- **自动密钥生成**：若用户未提供密钥，工具默认推荐并生成一个新的 RSA 4096 位私钥并保存。
- **内嵌资源释放**：程序内嵌了标准的 HA 数据库模板，运行时自动释放到当前目录，实现“零配置”启动。

## 输出物

### 1. HA 导入 SQL

按依赖顺序输出：

1. `users`
2. `profiles`
3. `profile_properties`

### 2. 迁移报告

至少包含：

- 导入用户数
- 导入角色数
- 自动补位角色数
- 被丢弃的 BS 玩家数
- 导入材质属性数
- 是否存在未签名材质属性
- bcrypt 直传用户数
- 被强制进入密码重置流程的用户数
- 冲突 / 跳过项

## 核心映射规则

## 1. 用户映射

BS `users.uid` 直接保留到 HA `users.uid`。

HA 新生成：

- `users.uuid`

直接迁移或尽量复用：

- `email`
- `password`
- `permission`
- `ip`
- `last_sign_at`
- `register_at`
- `verified`
- `remember_token`

默认值策略：

- `cbh = 1`
- `mbe = 0`
- `regip = ip`
- `mojang_uuid = NULL`
- `totp = NULL`
- `2FA = 0`

## 2. “BS 用户名”落地规则

`HA.users.username` 直接继承 BS 用户名。

针对当前支持的 BlessingSkin SQL dump 结构，这个值来自 `users.nickname`。

这里不再做：

- `nickname` 推导
- `email` 本地部分推导
- `uid` 补造用户名

实现要求：

1. 迁移器必须从 BS 输入中读取该用户的规范用户名；当前 dump 结构默认读取 `users.nickname`
2. 该值直接写入 HA `users.username`
3. 如果该用户缺失可恢复的 BS 用户名，该用户行应记入报告并按失败处理，而不是自动造名

## 3. 玩家到角色的映射

按你确认的意图执行：

1. 如果 BS 用户只有 1 个玩家，直接保留这个玩家
2. 如果 BS 用户有多个玩家，只保留“名字等于 BS 用户名”的那个玩家
3. 如果多个玩家里没有同名玩家，则视为“没有可用玩家”
4. 如果最终没有可用玩家，则自动创建一个补位角色，名字就是 BS 用户名

说明：

- 补位角色是迁移层行为，不回写 BS
- 被丢弃的玩家必须进入报告，不做静默忽略

## 4. 材质映射

BS 的 `textures.hash` 与材质文件名天然对应，一期直接复用该哈希。

迁移后：

- 材质文件复制到 HA 的材质目录
- `profile_properties.name = 'textures'`
- `profile_properties.value` 写入 base64 的 Yggdrasil textures payload
- payload 内的 URL 必须指向 HA 的 `/textures/{hash}`

`profiles.model` 规则：

- BS `steve` -> HA `default`
- BS `alex` -> HA `slim`
- 没有皮肤 -> `default`

## 5. 密码策略

你已经改成更简单的策略：

1. 迁移器从 SQL 中判断 BS 当前密码方法
2. 如果 BS 是 `BCRYPT`，则 `users.password` 直接透传
3. 如果 BS 不是 `BCRYPT`，则不做兼容校验迁移，统一进入密码重置流程

落库规则：

- `BCRYPT` -> 保留原哈希
- 非 `BCRYPT` -> 写固定失效标记 `BS2HA$RESET_REQUIRED`

这意味着：

- 一期不需要 HA 支持 multi-crypt
- 非 bcrypt 老用户迁移后不能直接沿用旧密码
- `security.password_cost` 只约束 HA 新生成或重置后的 bcrypt 密码

## 6. 签名要求

为了让迁移后的 Yggdrasil 资料立即可用，迁移工具应该使用 HA 私钥对 `textures` 属性签名。

未提供私钥时，可以保留一个显式的“未签名模式”用于联调，但不应作为生产默认方案。

## 推荐实现形态

推荐做成一个独立 CLI：

```text
bs2ha convert \
  --bs-sql ./bs.sql \
  --bs-textures ./textures \
  --ha-callback-url https://auth.example.com \
  --ha-private-key ./private_key.pem \
  --out-sql ./ha-import.sql \
  --out-report ./ha-import-report.json
```

## 实现分层建议

1. `parser`：解析 BS dump 中目标表的 `INSERT`
2. `staging`：构造成内部中间结构
3. `mapper`：执行用户、角色、材质、密码映射规则
4. `signer`：对 textures property 签名
5. `emitter`：输出 HA SQL
6. `reporter`：输出 JSON / Markdown 报告

## 当前边界

一期明确不做：

- 运行态 token 迁移
- OAuth2 客户端迁移
- 直接改写旧 BS 库
- 依赖旧 BS 线上接口补数据
