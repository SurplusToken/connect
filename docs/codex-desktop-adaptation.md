# Codex CLI / Desktop 高级适配方案

日期：2026-10-13。用户指定的 Codex++ 是 [BigPizzaV3/CodexPlusPlus](https://github.com/BigPizzaV3/CodexPlusPlus)。本文依据固定提交源码与当前官方文档；没有运行 Codex、Desktop 或参考应用，没有修改实际配置/认证/会话。上游源码与官方文档不自动等于本机 Desktop 内置版本的能力。

## 结论与分级

Connect 首版先做原生 Responses provider、认证交付、模型目录、配置备份与生效状态。CLI / Desktop 共用部分配置，但启动方式、环境变量与缓存不同，应作为两个能力目标。

| 级别 | Connect 能力 | 首版安排 |
| --- | --- | --- |
| 原生配置 | 独立 provider ID、模型、连接地址、版本适配、变更恢复 | 实现重点；当前 Codex 操作保护保持 |
| 模型目录 | 自定义模型可见性、上下文、推理等级、输入能力 | 与 provider 适配一起做，能力不明就不宣称支持 |
| 正常 Desktop 启动 | 配置持久有效，凭据对 Finder/开始菜单启动可用 | 永久模式必需；不以 shell 环境注入代替 |
| Desktop 独立实例 | 独立 Home、SQLite、GUI 用户数据与实例生命周期 | 第二阶段；按应用版本验证，不能只改 CODEX_HOME |
| CDP / UI 增强 | 注入页面、菜单扩展、模型列表辅助等 | 独立可选扩展，不作为标准接入必需依赖 |
| 会话 Provider 迁移 | 改写 rollout、状态库与索引 | 单独产品功能与事务，首版不自动执行 |

## BigPizzaV3/CodexPlusPlus 核查

固定版本：`45f1c1516ea208434731917690932be05c4a0586`；2026-10-13 获取，AGPL-3.0-only。源码在 `.references/CodexPlusPlus/`。

### 1. 当前架构是运行时增强

[README](https://github.com/BigPizzaV3/CodexPlusPlus/blob/45f1c1516ea208434731917690932be05c4a0586/README.md) 说明通过 CDP 与运行服务提供供应商、协议转换和 UI 增强，不修改官方 app.asar。不能把它归类为另外一些同名项目采用的安装目录补丁工具。

[`launcher.rs`](https://github.com/BigPizzaV3/CodexPlusPlus/blob/45f1c1516ea208434731917690932be05c4a0586/crates/codex-plus-core/src/launcher.rs) 可见 remote-debugging-port 启动链。与原生配置适配相比，CDP 增强增加了页面结构、启动参数和运行服务的版本依赖；Connect 无常驻代理的目标不能通过引入整套 runtime 自动实现。

其“官方登录”“官方登录 + API”“纯 API”“聚合”是不同模式，Chat Completions 会借助本地 Responses 转换。Connect 应从 SurplusToken 获得所需协议，而不是在本地转换。已有 ChatGPT 登录能否保留、哪些插件与云功能仍可用，需要分别验证，不根据一次模型请求成功作承诺。

### 2. 供应商切换具有快照与失败恢复

[`relay_switch.rs`](https://github.com/BigPizzaV3/CodexPlusPlus/blob/45f1c1516ea208434731917690932be05c4a0586/crates/codex-plus-core/src/relay_switch.rs) 的切换入口先读取管理器状态原件、捕获 live 配置/认证快照，保存前一个 provider 的状态，再应用目标；失败会分别恢复设置与 live 文件，并报告回滚失败。

[`provider_switch_isolation.rs`](https://github.com/BigPizzaV3/CodexPlusPlus/blob/45f1c1516ea208434731917690932be05c4a0586/crates/codex-plus-core/tests/provider_switch_isolation.rs) 专门覆盖切换第二个供应商后第一个配置被污染的问题。我们应借鉴“上一个配置”“目标配置”“live 当前值”三者分开的模型，不把当前盘上值直接回填到任意选中项。

源码存在这些测试不代表本轮运行过测试，也不代表整个切换流程已被我们完整审计。

### 3. 自定义模型目录比填一个模型字符串更重要

[`model_catalog.rs`](https://github.com/BigPizzaV3/CodexPlusPlus/blob/45f1c1516ea208434731917690932be05c4a0586/crates/codex-plus-core/src/model_catalog.rs) 读取 provider/model metadata 并生成目录。README 支持为模型标注自己的上下文窗口。Connect 应保存模型元数据的来源、版本和校验结果，不依据模型名字随意填 1M、图片能力或最高推理等级。

### 4. 多开隔离是多个状态根共同隔离

[`commands.rs`](https://github.com/BigPizzaV3/CodexPlusPlus/blob/45f1c1516ea208434731917690932be05c4a0586/apps/codex-plus-manager/src-tauri/src/commands.rs) 的 CodexMultiLaunchRequest / Worker / Instance 明确区分普通 launcher 单例与多开链路：实例拥有独立 CODEX_HOME、SQLite、Electron user-data-dir、CDP/helper 端口，并持久记录实例身份与运行状态。

这比 `CODEX_HOME=... open -a ...` 完整得多。但不能据此宣称官方所有 Desktop 版本都接受同样参数，尤其是应用壳、单实例锁和 Windows Store 路径发生变化时。Connect 原生独立实例模式要自己的版本验收；是否需要 CDP 是另一项能力。

### 5. Provider 同步不是普通 config 恢复

[`provider_sync.rs`](https://github.com/BigPizzaV3/CodexPlusPlus/blob/45f1c1516ea208434731917690932be05c4a0586/crates/codex-plus-data/src/provider_sync.rs) 包含会话扫描、同步锁、备份、删除墓碑等；[`provider-sync-flow.ts`](https://github.com/BigPizzaV3/CodexPlusPlus/blob/45f1c1516ea208434731917690932be05c4a0586/apps/codex-plus-manager/src/provider-sync-flow.ts) 将 UI 阶段划分为扫描、计划、备份、改写、更新索引、回滚、完成。

其目的涉及会话的 provider 归属和可见性，触及 sessions / archived_sessions 与状态索引。首版切换后如果旧会话暂时不可见，应显示解释和恢复原连接选项，不自动把会话重写成新 provider。将来实现迁移必须显示数量与范围、停机要求、独立备份和恢复入口。

许可证方面，首版只参考行为设计。直接抽取该 AGPL 工程源码之前需先确定本项目许可与分发方式；MIT 的 Magpie、CC Switch 和 Apache-2.0 的 tf-cli 更容易作为 Go 重实现的参考来源，但它们同样需要来源记录。

## 官方能力与本机验收的区别

当前[官方网关指南](https://learn.chatgpt.com/docs/enterprise/connect-to-a-gateway) 明确 CLI 与 macOS app 读取用户级配置，Windows 使用相应用户目录；Desktop 保存后需重启，凭据必须对 app 进程可用。用户级 provider 设置不能藏在项目 .codex/config.toml。

[配置参考](https://learn.chatgpt.com/docs/config-file/config-reference) 定义 model_catalog_json 启动时加载、provider 的认证/传输字段和受管配置边界。[认证文档](https://learn.chatgpt.com/docs/auth) 区分 file/keyring/auto/ephemeral，不应仅看到 auth.json 就假设它是唯一有效凭据来源。[环境变量](https://learn.chatgpt.com/docs/config-file/environment-variables) 列出 CODEX_HOME 和 CODEX_SQLITE_HOME，且 sqlite_home 配置优先于后者。

补充静态参考：OpenAI Codex main `669449b02c4b58de4bcd22485eb60c98a6e38a54` 的 [`app-server/src/model_catalog.rs`](https://github.com/openai/codex/blob/669449b02c4b58de4bcd22485eb60c98a6e38a54/codex-rs/app-server/src/model_catalog.rs) 在查询模型时检查当前 provider 要求，说明目录可见性也受受管配置约束。仅抓取了该文件；旧 core/auth 路径已变化，不据此编造当前认证存储实现。

版本探测建议：读取应用 Info.plist / Windows 包元数据与已登记可执行文件信息；静态无法确定内置 engine 版本时标记未知。未来解除运行限制后再用 CLI/engine 自身信息核实。官方 main、外部 CLI 和 Desktop 内置 app-server 必须分别记录，不能只显示一个“Codex 最新版”。

## CLI 临时适配

优先使用类似 tf-cli 的显式 -c 覆盖，Key 放只传给该子进程的专用环境变量。示例仅表示规划结构，不能直接用于真实模型验收：

```text
codex -c model_provider=surplustoken
      -c model_providers.surplustoken.name=SurplusToken
      -c model_providers.surplustoken.base_url=https://surplustoken.com/v1
      -c model_providers.surplustoken.wire_api=responses
      -c model_providers.surplustoken.env_key=SURPLUSTOKEN_CODEX_KEY
      -c model=<可用模型ID>
```

此处没有 Key 值。实施时严格按参数数组生成 TOML 值，不把字符串拼成 shell 命令。URL 路径由实际部署协议契约核实；用户面板 token 不作为推理 Key。

更高版本的侧挂 profile 文件可作为第二种策略，仍须按版本验证 profile 语义与清理。不要覆写用户已有同名 profile。默认不更改 sandbox、approval、MCP、skills、通知、遥测、已有历史目录。

## 永久配置的认证交付必须明确选择

| 策略 | 正常从系统启动 | 会否替换既有官方登录 | 卸载 Connect 后 |
| --- | --- | --- | --- |
| env_key + 子进程环境 | 仅从 Connect 拉起的新进程明确可用 | 通常无需改官方 auth store | 只写 config 不够，系统启动未必有变量 |
| 原生 API Key 文件/凭据库 | 目标版本支持并正确配置时可用 | 可能影响当前原生认证，需要备份与明确模式 | 只依赖客户端原生状态时可继续使用 |
| 独立凭据 helper | 安装 helper 后可适配非 shell 启动 | 可避免直接替换官方 auth 文件，仍须实测 | helper 必须留在安装路径，否则配置失效 |
| TOML 内直接 bearer | 不依赖外部环境 | 视 auth 模式而定 | 原理上可独立，但字段实验性且秘密落配置 |

推荐方向：普通永久模式优先验证目标版本的原生 API Key 凭据交付，满足“以后正常打开、卸载 Connect 仍可使用”；发现现有 ChatGPT 登录时显示“将切换原生认证”的独立变更，不无声覆盖。备用高级模式使用官方 command-backed provider auth + 系统钥匙串，但必须告知保留 helper 的依赖，不能宣称卸载后零依赖。

[官方指南](https://learn.chatgpt.com/docs/enterprise/connect-to-a-gateway) 支持 provider.auth 的命令 helper，也提醒 shell 环境不一定传给 Desktop。auth helper 与 env_key 等方案不能叠加。requires_openai_auth=true 会走 OpenAI 认证逻辑；不能把它与 env_key 混写然后猜测凭据优先级。

BigPizza 的纯 API 快照测试使用 OPENAI_API_KEY auth.json，并有官方混合 API 不应存该字段的校验；这是该项目的一种已实现策略，不是所有 Desktop 版本的兼容承诺。我们还未验证本机版本、系统钥匙串与直接登录状态共存，因此这仍是实现前必须用非真实目录夹具和后续实机明确的决策门槛。

不要默认设置机器全局环境变量或 launchctl setenv，也不要读取/复制现有 OAuth refresh token 来完成 SurplusToken 接入。官方登录到第三方网关的凭据共享是另一项功能。

## 模型目录与协议验证

1. 使用稳定且不冲突的 provider ID，例如 surplustoken；内置 provider 名不能随意替换。
2. 分组先匹配 Responses 协议，再给用户提供模型选择。/v1/models 的字符串列表不证明 streaming/tool calling/图片/推理能力。
3. 生成专属目录文件，使用绝对路径；目标 schema 与 engine 版本绑定。目录写入和 config 指针纳入同一可恢复事务。
4. 元数据至少包含真实模型 ID、显示名、上下文与压缩策略、允许的推理级别、输入模态、可用工具能力；哪些字段必需，以实际 schema 为准。
5. 只支持网关确认的能力，不默认复制 Magpie 模型条目里由本地代理模拟的 tool search 等功能。
6. SSE 文本、函数参数增量、工具结果回传、取消、认证错误、长上下文、WebSocket/压缩端点分别验证。协议转换由 SurplusToken 服务端负责；不兼容项在客户端卡片显示。
7. 应用后模型目录改变但旧进程还在运行时标“配置已写入，待重启”；模型可见、请求成功和配置写入成功是三个状态。

## Desktop 临时模式与高级入口

“从 Connect 启动”可在有已验收独立实例策略时提供：独立 Home + SQLite + GUI 用户数据。新实例可能没有原有会话和扩展，预览必须写清影响。目录按账号/连接身份稳定保存；删除实例必须排除用户原始 Home 与祖先路径。

普通永久模式修改共享用户级目录，对使用同一目录的 CLI 与 IDE 也可能生效；预览不能只写“修改 Desktop”。运行中应用不自动退出、重启或杀进程，提示用户自行重开。现阶段 Codex 操作保护仍保持，任何研究结果都不解除它。

建议 Codex 行点开详情提供：

- CLI / Desktop 安装来源与 engine 版本。
- 配置路径、认证策略、现有受管策略、模型目录版本。
- 已写入 / 待重启 / 未验证状态。
- 高级选项：推理等级、上下文元数据、独立实例（通过验收后）、恢复原连接。

不把会话迁移按钮挤进普通“应用配置”流程，也不自动迁移 SQLite 或 rollout。将来迁移需单独页面/向导：扫描摘要 → 备份范围 → 暂停写入 → 改写 → 更新索引 → 验证/恢复。

## 同名项目与其他对照

以下仅用于架构区别；用户指定的 Codex++ 仍是 BigPizzaV3。

- [ugarchance/codex-plus-plus](https://github.com/ugarchance/codex-plus-plus)，提交 b05ded31526690dec5c70c86a490e1a17a44a3a0，MIT：复制应用后补丁，私有用户数据/Home，原应用保持；home.cjs 用 staging 和数据库快照初始化，不是简单复制活动 SQLite 文件。
- [b-nnett/codex-plusplus](https://github.com/b-nnett/codex-plusplus)，提交 f98e7e9d1fa068dde9e0dddfb43b128acb4e2fd7，MIT：本地应用补丁、外部 tweaks runtime、更新 watcher。不能与 BigPizza 的“不修改 app.asar”说法混用。
- [Magpie codex.go](https://github.com/yetone/magpie/blob/a75fa9e54a7ec73a737681a3c345ae6fd0fe7d1f/internal/agent/codex.go)：模型目录、已有 provider/catalog 的恢复及用户新增模型字段保留值得参考；其官方登录旁路和订阅转发依赖网关，不直接采用。
- [CC Switch CodexStaleClientsNotice](https://github.com/farion1231/cc-switch/blob/d9a62313c8a276216ed40245c638753606831e95/src/components/providers/CodexStaleClientsNotice.tsx)：区分 daemon、Desktop/IDE、无法确认的状态；不能检测到进程就声称它已使用新模型。Connect 首版仅提示重启，不复制它的自动重启操作。

## 分阶段验收

A. 纯临时目录：TOML 顶层位置、未知字段/注释保留、认证策略互斥、目录 schema、已存在 provider 不覆盖、计划失效和多文件恢复。

B. 普通 CLI：解除禁止执行后，版本信息、只向子进程注入、流式工具调用和历史/权限设置保持。

C. Desktop 永久：正常系统启动获取凭据、模型选择器可见、重启后的新配置、恢复原连接、卸载 Connect 后持续工作。

D. 独立实例/CDP：主实例不受影响，Home/SQLite/GUI 状态相互独立，端口与单例锁不会复用到其他实例，应用升级后明确降级而不静默修改历史。

每阶段分别标记“静态验证”“夹具测试”“实机验证”；当前只有源码/文档调研，不宣称 B/C/D 已完成。
