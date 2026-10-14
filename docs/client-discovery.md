# 本机客户端发现

2026-10-13。安装检测与配置适配是独立状态；已安装不等于已经支持配置写入。

## 当前范围

检测 Codex CLI、Claude Code、oh-my-pi（omp）、OpenCode 和 Kimi（kimi）。按应用进程 PATH 优先，其次补查用户 .local/bin、Bun、npm、Volta、Homebrew、nvm、fnm 和版本管理器 shims 等已知目录。Windows 补查 AppData/npm、Scoop、WinGet、Node 目录，并识别 exe/cmd/bat。忽略相对 PATH 目录；不扫描项目目录、不执行用户 shell 初始化文件。

优先解析可执行文件对应的 npm package.json，校验包名及版本格式。没有包版本时，仅对非 Codex 客户端执行 --version：以用户主目录为工作目录、关闭 stdin、补齐运行时 PATH，超时 3 秒、输出最多保留 32 KiB。检测并行执行，结果仍按固定客户端顺序展示。无法识别、执行失败和超时分别保留原因，不冒充“未安装”。

Windows cmd/bat 包装器没有可靠包元数据时显示版本未知，不借助 cmd.exe 或 PowerShell 执行任意包装脚本；exe 版本探测隐藏控制台窗口。Windows 发现模块已交叉编译，尚未实机验收。

Codex 保持保护：只查可执行文件及 npm 包版本，绝不执行 codex --version；macOS 的独立 Codex.app 只查标准应用目录和 XML/二进制 Info.plist 版本，不启动应用，不读 CODEX_HOME、认证或会话。同时发现 ChatGPT.app 时展示其 Desktop 应用版本；不探查内嵌 Codex 引擎或凭据。非标准应用目录可能显示未知。

界面显示版本与安装状态，点击版本号打开 Kumo Popover，查看位置、版本来源和失败原因。omp/Kimi 即使安装成功，配置适配仍显示待完成。OpenCode 已安装但版本无法确认，或者主版本不是 2，永久配置继续关闭。

## 参考与取舍

- tf-cli：参考 internal/process/command_windows.go 对 npm 包装器和运行时的区分，避免在 shell 中拼接执行字符串。
- Magpie：参考 internal/proc/find.go、version.go 的 GUI 路径补查和包元数据优先策略；此轮独立实现发现模块，没有复制其完整进程、安装或更新系统。
- CC Switch：参考 src-tauri/src/commands/misc.rs 的 PATH 默认安装、版本来源与平台探测区分；未复制 AGPL 代码，不读取其 Codex 会话数据库版本。

当前返回首个匹配命令，没有完整枚举同一工具的多份安装。版本管理器目录属于有限补查，不能保证与交互 shell 的 alias/default 完全一致；结果详情中的真实路径用于确认。尚未增加安装、升级、自动修改 PATH、WSL 扫描或安装器下载。

测试覆盖 PATH 优先级、非可执行文件、包名验证、Windows npm shim、GUI 补查、超时原因、输出上限与 Codex 零执行保护。本机真实发现结果另外通过网页及原生窗口核对。
