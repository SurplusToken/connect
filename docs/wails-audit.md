# Wails 桌面外壳排查

日期：2026-10-13。对象：Wails v3.0.0-beta.28、当前 Kumo UI 预览、本机 macOS 27.0.1 / arm64。结论不是跨平台认证；未运行 Codex 或修改其配置。

## 发现与处理

| 项目 | 结论 | 处理与证据 |
| --- | --- | --- |
| 根页面滚动与回弹 | 实际缺陷，前一轮已修复 | html/body/#app 限高、禁止根滚动；main 独立滚动且禁用 overscroll。原生窗口上下滚动验证。 |
| Wails runtime 初始化 | 实际遗漏，本轮修复 | 新增与 Go 完全同版本的 @wailsio/runtime。原生 wails: / wails.localhost 环境先初始化再挂载 React；普通浏览器预览不调用原生桥。官方要求即使不调用 API 也初始化 runtime。 |
| 空白区浏览器右键菜单 | 桌面行为缺陷，本轮修复 | 使用 Wails 的 --default-contextmenu 控制；外壳 hide，输入框 auto。原生空白区右键无 Reload，邮箱输入框仍有 Cut/Copy/Paste。 |
| 初始化失败无提示 | 本轮补齐 | 启动失败时显示中文提示，避免只留下空白页。 |
| 最低系统版本描述 | 证据不足，已修正文档 | 13.0 是 Go/链接器部署目标，不等于 UI 最低版本验收。Kumo 含 oklch、color-mix、@property；产物已转换 light-dark，不能仅根据源码判断支持范围。 |
| macOS 发行签名 | 尚未完成 | 当前只有 Go 链接器生成的 ad-hoc 签名：无 TeamIdentifier，Info.plist 未绑定，资源未封签。不是可对外发行的 Developer ID 签名包。需要签名、公证与 staple 流程。 |
| Windows 打包 | 尚未完成 | 无 Windows 图标/manifest/安装器/WebView2 安装检测与实机验证。README 的 go build 命令不能代表发行流程完成。 |
| Intel Mac | 尚未验证 | 当前可执行文件仅 arm64；没有 Universal 包。 |
| 原生菜单中文化 | 产品完成度缺项 | 默认 File/Edit/View 等仍为英文；不是 Wails 强制规则。未来按中文产品需求定制并保留系统编辑角色与快捷键。 |
| 单实例 | 后续业务设计 | UI 原型不要求单实例；接入配置写入后，应使用单实例或写入锁防止多个窗口同时改文件。当前没有写文件服务。 |

## 已检查、未发现配置问题

- 使用 production 构建标签；未启用 devtools 或 private_mac_apis。
- main.go 使用官方 AssetFileServerFS 与 embed.FS，资源来自本地打包产物，入口为 /；不依赖 Vite 开发服务器运行。
- 保留系统原生标题栏，因此不需要自定义 --wails-draggable 区域；不要在表单或整页设置拖窗区域。
- 窗口 1000×650、最小 900×600；内容区 min-height:0，标题栏不收缩，侧栏固定。窄浏览器预览另有折叠布局。
- 页面声明 zh-CN、viewport 与 light color-scheme；Dialog 使用 Kumo Portal，根节点 isolation 保留。
- 图标运行时 PNG 与 ICNS 使用同一源图和构图；bundle executable / icon 路径匹配，Info.plist 语法有效。
- macOS CGO 部署参数与 LSMinimumSystemVersion 同为 13.0。这只能验证元数据一致，不能代替旧系统运行。
- 关闭最后窗口退出是本产品当前明确配置；不是泄漏或框架错误。
- 自定义 Makefile 可以构建 Wails，Taskfile / wails3 CLI 是官方工具链便利入口，并非所有工程必须采用。
- 当前没有注册 Go service，没有外链导航入口，也没有凭据持久化、子进程启动或后台网络任务；对应业务正确性无法通过审计 UI 外壳证明。

## 接入后端前应完成

1. 登录、网络访问、客户端配置与启动放入明确的 Go 服务，输入校验、超时、取消与错误脱敏在 Go 边界执行。
2. 密码仅作登录输入；保存 token/key 时使用系统凭据存储或明确受限文件，不放 localStorage、日志和命令行。
3. 写入备份、原子替换、外部修改检测、恢复和进程生命周期必须独立验收。
4. 外部网站使用 Wails Browser API 打开；不要让带原生桥的应用页面跳转任意远程站点。
5. 为最低支持 macOS/WebKit 和 Windows/WebView2 建立实机矩阵，验证弹窗、键盘焦点、缩放、滚动、复制粘贴与图标。
6. 发布前补齐平台安装包、签名公证、版本与升级方案；不把本机预览包标作正式发布。

## 本次验证

- npm 安装锁定 runtime 3.0.0-beta.28，依赖审计报告 0 vulnerabilities。
- make package-mac 成功：TypeScript 检查、Vite 构建、Go production 编译及 .app 资源复制。
- go vet ./... 通过；plutil -lint 通过。
- 本机重新启动 Connect，原生页面加载正常；右键菜单验证通过。
- 未进行 Windows、Intel Mac 或 macOS 13 实机验证；没有实际后端请求测试。
- 未以“没有编译错误”替代上述平台与业务验收。

## 官方依据

- [Frontend Runtime](https://v3.wails.io/reference/frontend-runtime/)：runtime 初始化及副作用导入。
- [Menus](https://v3.wails.io/guides/menus/)：上下文菜单 CSS 控制。
- [Overscroll](https://wails.io/docs/next/guides/overscroll/)：根页面溢出处理（该页属于旧版文档，方案已在当前 v3 原生窗口验证）。
- [macOS Packaging](https://v3.wails.io/guides/build/macos/)：bundle、公开 API、签名与分发。
- [Building Applications](https://v3.wails.io/guides/build/building/)：官方构建工具链。
- [WebKit Safari 17.5](https://webkit.org/blog/15383/webkit-features-in-safari-17-5/)：light-dark 的版本背景；当前构建产物已转换它，不据此直接设最低系统版本。
