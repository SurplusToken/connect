# Prior work 与可复用组件

调研日期：2026-10-13；实现状态更新：2026-10-14。当前实现以 [API 与配置接入](api-configuration-implementation.md)、`go.mod` 和 `frontend/package-lock.json` 为准。来源为 GitHub 搜索、项目 README、公开源码与官方文档；未安装或运行这些项目，后续已对 SurplusToken 测试账号执行登录与只读接口验证，结果见下文。除已固定提交的 tf-cli 外，链接指向上游当前分支，落地前应固定依赖版本并复查许可证。

## 结论

优先组合：**tf-cli 的启动方案设计 + CC Switch 的配置管理经验 + 直接调用 Sub2API 登录与 Key 管理 API + Wails v3 外壳**。不建议直接 fork 一个常驻代理工具作为产品核心。

本轮搜索没有确认一个同时完成 Sub2API 用户授权、自动 Key 接入、五类目标客户端、双模式与无常驻代理的现成项目。该结论限于本轮核查范围，不是对 GitHub 全量项目的保证。

## 最值得研究的项目

| 项目 | 核实到的能力 | 本项目的复用方式与边界 | 许可证 |
| --- | --- | --- | --- |
| [TokenFlux/tf-cli](https://github.com/TokenFlux/tf-cli) | Go；Claude Code、Codex、OpenCode、Pi 进程注入；网页导入、模型探测、Windows 进程处理 | 首选参考。提取纯逻辑并保留测试；没有永久配置、Kimi 与 oh-my-pi 的现成完整适配 | Apache-2.0 |
| [farion1231/cc-switch](https://github.com/farion1231/cc-switch) | Tauri/Rust 桌面工具；provider 配置、备份、凭据切换、直连及路由模式 | 借鉴配置合并、原始状态备份、恢复和冲突 UI；Rust 实现不能直接作为 Go 包导入 | MIT |
| [PalenaAI/claude-code-auth-helper](https://github.com/PalenaAI/claude-code-auth-helper) | Go；浏览器 SSO、OIDC、凭据 helper、系统钥匙串；支持令牌直通/交换 | 授权体验和凭据生命周期很接近；网关须支持其身份验证/交换方式，不能直接当作 Sub2API 静态 Key 创建器 | Apache-2.0 |
| [daodao97/code-switch](https://github.com/daodao97/code-switch) | Go + Wails **v3**；Claude/Codex provider 管理 | 参考 Go/桌面界面组织。它将客户端指向本地代理；可参考 v3 的组织方式，具体 API 需按本项目锁定版本核对，不能整体套用 | Apache-2.0 |
| [Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api) | 用户认证、Key 和分组管理的实际上游 | 用作 HTTP 契约来源；核实部署版本，不直接引入整个服务端 | LGPL-3.0（GitHub 标识） |
| [BigPizzaV3/CodexPlusPlus](https://github.com/BigPizzaV3/CodexPlusPlus) | Codex 桌面增强工具 | 作为桌面兼容性调查入口；本轮未验证其内部实现，不选为直接代码依赖 | AGPL-3.0 |

### tf-cli 源码入口

本地：`.references/tf-cli/`；固定提交：`d32cefdda83201aa92655083637b3e27e20cbf1d`。

- [`internal/harness/inject.go`](https://github.com/TokenFlux/tf-cli/blob/d32cefdda83201aa92655083637b3e27e20cbf1d/internal/harness/inject.go)：`Input` / `Plan`、环境合并、客户端启动配方和 Pi 临时扩展。适合抽出无 UI 的适配层。
- [`internal/gateway/`](https://github.com/TokenFlux/tf-cli/tree/d32cefdda83201aa92655083637b3e27e20cbf1d/internal/gateway)：模型目录、协议探测、响应分类和相关测试。
- [`internal/process/`](https://github.com/TokenFlux/tf-cli/tree/d32cefdda83201aa92655083637b3e27e20cbf1d/internal/process)：Windows/npm 可执行文件发现与进程处理。
- [`internal/cli/web_import.go`](https://github.com/TokenFlux/tf-cli/blob/d32cefdda83201aa92655083637b3e27e20cbf1d/internal/cli/web_import.go) 与 [Web Import 协议](https://github.com/TokenFlux/tf-cli/blob/d32cefdda83201aa92655083637b3e27e20cbf1d/docs/integrations/web-import.md)：短时本地监听、Origin、会话验证与网页导入状态。

Go 的 `internal` 包不能被本项目直接跨仓库 import。实际复用需抽取并归属到自己的包，或推动上游发布公共包；不要把参考目录当成构建依赖。该提交 `go.mod` 声明 Go 1.27.0，代码移植时另行确认最低工具链要求。

首版已决定不采用网页导入，以下仅保留为上游研究笔记。上游允许用户确认“未验证”的网页导入。上游协议的 HTTP `202` 只代表本地确认，不代表 Key 校验和保存成功，GUI 应区分这些阶段。

### CC Switch 源码入口

- [Codex 配置](https://github.com/farion1231/cc-switch/blob/main/src-tauri/src/codex_config.rs)
- [OpenCode 配置](https://github.com/farion1231/cc-switch/blob/main/src-tauri/src/opencode_config.rs)
- [通用配置](https://github.com/farion1231/cc-switch/blob/main/src-tauri/src/config.rs)
- [数据库备份](https://github.com/farion1231/cc-switch/blob/main/src-tauri/src/database/backup.rs)

本轮核实了上述文件存在和 README 中的备份/状态机制，尚未逐函数审计。尤其要区分它自身数据库备份与客户端原始配置备份。可优先借鉴：首次改写保留原件、记录最后写入状态、切回原认证模式、显示备份位置；不把其代理/MCP/Skills 管理范围带入首版。

### Sub2API 授权边界

[用户路由源码](https://github.com/Wei-Shaw/sub2api/blob/main/backend/internal/server/routes/user.go) 显示 Key 路由受用户 JWT 中间件保护，包含 `/keys` 的 GET/POST、`/keys/:id` 的 GET/PUT/DELETE，以及 `/groups/available`。这些是传入 v1 路由组下的相对路径；完整部署前缀、字段、返回明文策略和权限必须用实际部署核实。

这证明上游具有用户 Key 管理能力，**不能证明 SurplusToken 已具备桌面授权或 TokenFlux Web Import**。当前决策是在 Connect 中输入账号密码并直接请求登录 API，面板 access token 仅用于用户 API，默认保存在内存；不使用管理员 Key，不把面板 token 交给客户端。TokenFlux 与 Sub2API 的代码血缘本轮未独立确认，不作为接口兼容依据。

## 客户端一手资料

| 资料 | 应核实的适配问题 |
| --- | --- |
| [Codex 官方配置](https://learn.chatgpt.com/docs/config-file/config-reference) | `model_providers`、`base_url`、`env_key` 与认证方式；当前 `wire_api` 支持 Responses。桌面启动与认证需独立实测 |
| [oh-my-pi](https://github.com/can1357/oh-my-pi#custom-openai-compatible-providers) | 当前 README 使用 YAML：`~/.omp/agent/models.yml` / `config.yml`；不能把 tf-cli 的 `pi` 当作 `omp` |
| [OpenCode](https://github.com/anomalyco/opencode) | 目标版本的配置层级、provider 配置和临时注入接口 |
| [Kimi Code](https://github.com/MoonshotAI/kimi-code) / [配置文档](https://moonshotai.github.io/kimi-code/en/configuration/config-files) | 新项目、TOML、`KIMI_CODE_HOME`、自定义 provider 与模型能力；旧 [kimi-cli](https://github.com/MoonshotAI/kimi-cli) 已归档 |

## 可复用组件候选

下表区分已采用组件与后续候选；实际版本以依赖清单为准。优先标准库，只有确定需求后才引入库。

| 组件 | 用途 | 采用建议 / 限制 | 许可证 |
| --- | --- | --- | --- |
| [Wails v3](https://github.com/wailsapp/wails) / [文档](https://v3.wails.io/) | 桌面窗口、Go/前端 Service 桥接、平台打包 | 已采用；Go 模块锁定 v3.0.0-alpha.63，前端 @wailsio/runtime 锁定 3.0.0-beta.28。参考 v3 文档时核对锁定版本的 API | MIT |
| [zalando/go-keyring](https://github.com/zalando/go-keyring) | macOS/Windows 系统凭据库 | 已用于非 macOS 系统凭据库；macOS 使用 go-keychain。处理失败与拒绝访问，Windows/Linux 尚未实机验收 | MIT |
| [adrg/xdg](https://github.com/adrg/xdg) | 跨平台应用目录 | 可选；简单需求也可用 `os.UserConfigDir`，客户工具目录仍由适配器决定 | MIT |
| [pelletier/go-toml/v2](https://github.com/pelletier/go-toml/tree/v2) | Codex/Kimi TOML 解析与验证 | 优先候选；普通反序列化/序列化不等于无损编辑，注释与未知字段需要专项处理 | MIT（已读 LICENSE） |
| [tidwall/jsonc](https://github.com/tidwall/jsonc) | 接受注释、尾随逗号的 JSON 输入 | 已采用；只解决读取，局部编辑由 internal/configedit 的 Magpie 改编算法完成 | MIT |
| [microsoft/node-jsonc-parser](https://github.com/microsoft/node-jsonc-parser) | JSONC 语法树与局部修改 | 若评估 JS 配置编辑可参考；Go 核心不因此引入常驻 Node 依赖 | MIT |
| [gofrs/flock](https://github.com/gofrs/flock) | 进程间文件锁 | 已用于配置引擎的跨实例写锁；外部客户端不一定遵循同一把锁，仍需写前比对 | BSD-3-Clause |
| [natefinch/atomic](https://github.com/natefinch/atomic) | 跨平台单文件替换 | 作为实现候选；验证 Windows 文件占用、权限和故障场景，不代替事务或备份 | MIT |
| [spf13/cobra](https://github.com/spf13/cobra) | CLI 子命令、补全 | CLI 规模增长后采用，少量命令可先用标准库 | Apache-2.0 |
| [creack/pty](https://github.com/creack/pty) + [UserExistsError/conpty](https://github.com/UserExistsError/conpty) | Unix PTY / Windows ConPTY | 仅需要接管终端时考虑；首版外部终端可减少这部分复杂度 | MIT |

oh-my-pi 还需要 YAML 的保留字段/注释编辑方案，本轮未确定库。配置写入应独立做小型验证，不因存在 TOML/JSON/YAML parser 就假定能无损往返。HTTP、超时、随机数和进程执行优先使用 Go 标准库以及 Wails runtime。

## 搜索到但不适合作为主干的方案

- [mise42/sub2api-cli](https://github.com/mise42/sub2api-cli)：导入上游 OpenAI OAuth 账号，使用 Sub2API 管理员接口；不是给终端用户分发推理 Key 的助手。
- [yunhai-dev/sub2api-cli](https://github.com/yunhai-dev/sub2api-cli)：管理员 API CLI，范围和身份权限不匹配。
- [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI)：Go 代理服务，核心职责与本项目无常驻代理的要求不匹配。
- [mos1128/ccg-gateway](https://github.com/mos1128/ccg-gateway)：网关与效率工具，可继续研究客户端配置体验；不应从项目名称推断具备 Sub2API 用户 SSO。

## 后续验证与已完成项

1. Claude Code / OpenCode v2 的规划、写入与恢复已实现并通过隔离目录联调；继续验收真实客户端推理和 Windows 文件行为。
2. Codex CLI 与 Desktop 目前禁止启动/重启及运行测试，只做静态适配验证。解除限制后再验证认证、模型选择、请求地址和卸载后的独立使用。
3. 直接 API 登录、Key 管理与原生登录恢复已接入；真实 TOTP 仍待验收，暂不采用网页回调。
4. 验证 Windows GUI 拉起 npm 安装的交互式 CLI；确认临时配置不会被现有用户配置覆盖。
5. 根据实际复用文件整理第三方声明，再决定项目自身许可证。


## 初次直接 API 登录验证（历史记录）

2026-10-13 在 `https://surplustoken.com` 使用用户提供的测试账号执行：

| 请求 | 结果 |
| --- | --- |
| `POST /api/v1/auth/login` | 200，返回 access token、refresh token、过期时间及用户对象 |
| `GET /api/v1/user/profile` | 200 |
| `GET /api/v1/keys?page=1&page_size=20` | 200，分页列表为空 |
| `GET /api/v1/groups/available` | 200，可见 5 个分组 |

以下仅描述初次只读验证；后续 Key 创建、模型目录和原生登录恢复结果见 [当前实现与验证边界](api-configuration-implementation.md#验证边界)。当次认证信息仅在请求进程内存中使用，输出仅保留状态码与响应结构。未创建/删除 Key，未测试刷新/二次验证，未发起推理请求。登录本身可能产生服务端登录记录。账号密码与返回令牌不写入文档或仓库。

## 客户端目录变量与隔离边界

- **Codex：`CODEX_HOME`**。官方文档定义为配置、认证、日志、会话等状态根目录，默认 `~/.codex`，覆盖目录必须已存在。另有 `CODEX_SQLITE_HOME`，且配置中的 `sqlite_home` 优先，说明单一 Home 变量不保证覆盖所有状态。仅核查文档，未启动 CLI/Desktop。来源：[官方环境变量](https://learn.chatgpt.com/docs/config-file/environment-variables)。
- **Claude Code：`CLAUDE_CONFIG_DIR`**。覆盖用户设置、会话历史和插件目录；项目与受管设置仍有自己的作用域。不能据此断言系统钥匙串也隔离。来源：[官方设置文档](https://code.claude.com/docs/en/settings)。
- **Kimi Code：`KIMI_CODE_HOME`**。默认 `~/.kimi-code`，修改后读取该目录下 `config.toml`，而不是只换单个文件路径。来源：[配置文档](https://moonshotai.github.io/kimi-code/en/configuration/config-files)。
- **oh-my-pi：`PI_CODING_AGENT_DIR`**。当前仍使用这个名字，不应自行假定为 `OMP_HOME`。默认 profile 下覆盖原生 agent 数据目录，但通用 `getConfigDirs()` / `findConfigFile()` 的 OMP 根目录由 `PI_CONFIG_DIR` 控制；命名 profile（`--profile`、`OMP_PROFILE` 或 `PI_PROFILE`）会忽略 `PI_CODING_AGENT_DIR`。外部工具配置发现仍可能访问其他用户目录。来源：[上游配置解析文档](https://github.com/can1357/oh-my-pi/blob/main/docs/config-usage.md)、[上下文发现文档](https://github.com/can1357/oh-my-pi/blob/main/docs/context-files.md)。
- **OpenCode v1 调研记录（不作为当前 v2 接口依据）：`OPENCODE_CONFIG_DIR` / `OPENCODE_CONFIG` / `OPENCODE_CONFIG_CONTENT`**。配置按来源合并，不是全部替换。官方 CLI 文档列出凭据在 `~/.local/share/opencode/auth.json`，不能把 `OPENCODE_CONFIG_DIR` 当作凭据与会话根目录。数据目录/XDG 的平台实现仍需进一步源码确认。来源：[CLI 环境变量与认证](https://opencode.ai/docs/cli/)、[配置优先级](https://opencode.ai/docs/config/)。

建议将“进程配置注入”和“独立配置目录”作为适配器的两种能力：前者尽可能保留用户已有会话/插件，后者便于独立账号及测试，但可能改变历史记录和扩展发现。不要通过重设整个操作系统 `HOME` 来伪装隔离，也不要将目录隔离等同于安全沙箱。


当前只开放 OpenCode v2；目录覆盖、多层合并和原生 providers 格式见 [OpenCode v2 迁移](opencode-v2-migration.md)。

## 当前桌面实现（2026-10-14）

工程采用 **Wails v3**；Go 模块为 `v3.0.0-alpha.63`，前端 runtime 为 `3.0.0-beta.28`。两者版本号当前不同，依赖清单分别记录各自的锁定版本。桌面入口使用单窗口、本地静态资源和 `httpapi.Bridge` Go Service，复用开发 API 的请求处理逻辑。版本升级时需核对桥接 API，并重新验证原生调用；不能仅用前端构建结果判断兼容性。

双模式建议：临时启动使用 Connect 管理的持久独立目录，以保留会话；每次只向子进程注入连接配置，退出后清理含秘密的临时文件。永久模式写入实际生效目录，保存原文件备份、修改字段与写入后摘要，恢复前再次检查文件，遇到外部修改则展示冲突。两者共用同一份客户端/Key/模型选择，执行适配器分开。当前已实现 Claude Code / OpenCode v2 的永久配置、备份与恢复；临时启动尚未接入，Codex 保持禁止操作。

## 配置流程与 Codex 高级适配补充

2026-10-13 已对 tf-cli、CC Switch、Magpie 和用户指定的 BigPizzaV3/CodexPlusPlus 固定提交进行进一步静态源码调查：

- [配置修改、备份与恢复流程](configuration-workflows.md)：界面入口、字段所有权、事务、冲突、恢复与验收。
- [Codex CLI / Desktop 高级适配](codex-desktop-adaptation.md)：模型目录、认证策略、正常启动、隔离实例、CDP 与会话迁移边界。

这两篇是接入阶段的设计依据，不是当前实现的完整清单；实际范围见 [API 与配置接入](api-configuration-implementation.md)。早期只读核查结果保留为带日期的历史记录。参考源码调研没有运行参考应用或启动/修改 Codex。
