# API 与配置接入：当前实现

日期：2026-10-13。真实网关固定 `https://surplustoken.com`；无需网页授权回调。以下明确区分已实现、已验证与暂未开放。

## 真实 API

- Go 调用用户登录、TOTP 二次登录、令牌刷新、退出、用户资料、可用分组、分页 Key 列表、单 Key 读取与创建。
- 推理凭据仅用于 `/v1/models`，不会把面板 access token 交给模型端点或客户端。
- HTTP 限时、响应大小限制、拒绝重定向；错误响应不原样回传，已映射余额不足等原因。
- 面板 access token 仅保存在 Go 内存；原生 refresh token 保存到系统凭据库，重启时刷新换取新登录态。密码不落盘。网页开发预览在 localStorage 记住本地会话 ID 和账号名；服务重启或会话超过 8 小时需重新登录，不保存网关令牌。
- 前端只得到用户 DTO、分组、Key 掩码和模型 ID；Key 明文、计划原始字节与备份内容不进入 API 的公开响应。
- 本地 API 要求专用请求头和登录会话，并限制 Origin；不开放 CORS。原生窗口通过 Wails Service 桥接相同处理逻辑，开发服务只监听 127.0.0.1:43129。
- TOTP、刷新、重定向拒绝通过 HTTP 夹具测试；真实 TOTP 尚未验收，真实刷新已通过原生应用关闭/重开登录恢复验证。

## 已开放配置

| 客户端 | 状态 | 配置内容 |
| --- | --- | --- |
| Claude Code | 永久写入/恢复已实现 | CLAUDE_CONFIG_DIR 或默认 settings.json；连接、Token、主模型和辅助模型环境字段；分组必须声明 Messages 兼容 |
| OpenCode v2 | 永久写入/恢复已实现 | OPENCODE_CONFIG_DIR / XDG 用户目录；providers.surplustoken、原生 package/settings、模型目录；不改默认模型或标题模型 |
| Codex CLI / Desktop | 保持禁止操作 | 不执行、不读取认证、不修改目录；写入入口拒绝 codex |
| oh-my-pi / Kimi | 待适配 | 只展示入口，不允许选择写入 |
| OpenCode v1 | 停止适配 | 当前仅开放 v2，不向 v1 写入原生 v2 格式 |
| 临时终端启动 | 待接入 | UI 明示后续接入，执行按钮禁用 |

客户端版本优先从对应 npm 包元数据读取，非 Codex 的缺失版本再用限时 --version 探测；未运行任何 Codex 命令。未安装时可预先配置 OpenCode v2；无法确认已安装版本时不开放写入。GUI 的 PATH 缺失现在补查已知安装目录，完整多安装和版本管理器默认解析尚待扩展。详见 [本机客户端发现](client-discovery.md)。

Claude 现有 Bedrock/Vertex/Foundry 路由打开时阻止规划；OpenCode 已禁用 surplustoken 时阻止规划，已有 enabled_providers 列表则在预览中显式加入此项。项目作用域、受管策略和用户 shell 的覆盖并未被完全枚举；配置写入成功不等于所有项目的最终请求路由已验证。

## 写入与恢复

1. 从后端登记的客户端路径读取，纯内存生成字段级 JSON/JSONC 补丁。
2. 前端预览只含路径和脱敏字段前后值，计划 ID 绑定本次登录会话，有效期 10 分钟。
3. 应用时拿进程内互斥与跨实例文件锁，再次检查当前文件摘要和存在性；预览后变化会阻止提交。
4. 操作前原始字节、写入后字节、字段范围、文件存在性和 pending 状态保存为私密备份；随后同目录暂存、刷盘、rename、完成记录。
5. 当前每个客户端事务只修改一个文件，多客户端独立执行并逐项显示成功/失败。尚未宣称通用多文件事务已实现。
6. 启动后读取记录时检查 pending：目标匹配 after 则补记完成，匹配 before 则记未提交，其余标冲突，不覆盖现场。
7. 恢复先检查受管字段当前值：仍等于 applied 才恢复 original；被外部改过则拒绝。其他字段继续保留。
8. 没有外部改动时恢复原件字节；原来不存在且内容仍完全是 Connect 所写时删除新文件。恢复本身也先备份，且有独立预览计划。

本轮不自动清理历史，因此第一次应用前原件会保留；后续保留策略、完整快照还原、复杂冲突逐字段选择尚未实现。符号链接路径目前明确拒绝，不声称链接/硬链接无损支持。Windows rename、ACL 和文件占用尚未实机验证。

历史目前只加载与当前登记配置路径一致的记录；更换配置目录后旧备份仍在磁盘保留，但不会显示或自动恢复到旧路径。跨目录备份浏览尚待实现。

备份目录为 os.UserConfigDir 下的 SurplusToken Connect/backups。macOS 目录 0700、文件 0600；客户端含秘密文件也用 0600。备份是受限的本地明文（JSON 字节以 base64 表示，不是加密），不应分享或提交。API Key 备份还未实现加密导出；系统凭据库目前只存面板 refresh token。

## 验证边界

- 真实测试账号：登录、用户资料、五个分组、Key 列表、经用户确认创建一个名为 SurplusToken Connect 的 Key。
- `/v1/models` 初次返回 INSUFFICIENT_BALANCE；原因已展示。随后在原生 Connect 中请求成功，OpenCode 模型菜单实际显示 gpt-5.5。未发起模型推理，也未自动补充余额。
- 配置测试与 UI 应用/恢复均使用显式测试目录，真实客户端文件保持不变。
- Go race 测试、go vet、前端类型检查与构建、macOS 打包通过。解锁后已用最新原生包验证 Wails Service 登录、用户资料弹窗、Key 列表、客户端检测和真实模型菜单；原生配置预览操作因用户接管窗口而停止，未点击应用写入。配置应用与恢复的完整 UI 联调仍以隔离目录网页测试为依据。
- 覆盖注释/CRLF、秘密不出现在 Plan/History、跨会话计划拒绝、计划过期/外部修改、恢复保留未知字段、恢复冲突、备份失败、pending 恢复、无原文件恢复、符号链接拒绝、Codex 保护。
- 未验证真实 Claude/OpenCode 推理、Desktop 凭据交付、临时终端生命周期、Windows 平台。

## 文件入口

- internal/gateway：网关 HTTP 契约与令牌生命周期。
- internal/httpapi：本地 UI API、身份检查、Key/模型/适配器连接。
- internal/configuration：已登记路径、规划、备份、应用、记录与恢复。
- internal/configedit：从 Magpie 改编的纯 JSON/JSONC 补丁算法，MIT 许可保留。
- frontend/src/api.ts：本地会话、开发预览恢复、API 请求和错误处理。

## 保持登录

- 原生应用默认保持登录，独立服务标识为 com.surplustoken.connect.session。macOS 直接调用 Security.framework，不将秘密放入命令参数。系统钥匙串保存失败会明确提示，本次登录仍可使用，不降级到明文文件。
- 令牌轮换同步更新保存值；旧本地会话不能覆盖新登录凭据。启动恢复串行化，避免重复消费刷新令牌。
- 网络失败保留凭据并允许重试；401/403 拒绝时清理失效状态。退出先清除本机保存值，再尝试撤销远端刷新令牌；远端撤销失败不妨碍本机退出。不会删除推理 API Key。
- 测试覆盖保存、跨实例恢复、轮换写回、秘密不进入 UI 响应、退出清除、临时失败保留、失效清理。Windows/Linux 系统凭据库尚未实机验证。
- cmd/connect-dev：开发 API、真实网关测试目录模式及显式离线夹具模式。离线夹具不编译到 Wails 应用。

客户端列表只显示单独配置按钮，Key 在弹窗中选择。OpenCode 按供应商模板接入，不指定默认模型或标题模型；Claude Code 的模型仍在弹窗确认。OpenCode v2 原生字段已通过安装的 2.0.28 CLI 在隔离 HOME 中以 models --standalone 加载验证，没有推理请求。混合协议分组暂不开放 OpenCode 配置，避免猜测模型协议。

OpenAI 模板位于 internal/httpapi/templates/opencode-openai.json：兼容 Responses 运行时、canonical openai、HTTP 传输和固定网关地址。Key 与模型目录由后端填入。模板不修改 model、small_model 或 agents.title.model，已有设置原样保留。默认预览展示模板与模型数量，字段级脱敏差异通过 Kumo Collapsible 展开。
