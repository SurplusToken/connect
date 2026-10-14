# 第三方资源声明

- Lucide 1.55.0：通用操作图标，ISC 许可。来源：https://lucide.dev/ 。许可副本：`frontend/public/licenses/lucide.txt`。
- Lobe Icons static SVG 1.95.1：OpenAI、Claude、OpenCode、Kimi 图标，MIT 许可。来源：https://github.com/lobehub/lobe-icons 。下载地址格式：`https://unpkg.com/@lobehub/icons-static-svg@1.95.1/icons/{name}.svg`。许可副本：`frontend/public/licenses/lobe-icons.txt`。

Claude 与 Kimi 使用 `-color` 版本，保留 SVG 原始色彩；Kimi 白色字形配深色底。OpenAI 与 OpenCode 保留资源原色。oh-my-pi 暂用 Lucide Terminal 通用图标，不冒充品牌标识。品牌名称和商标归各自所有者所有。

## Kumo 版本新增依赖

- Kumo 2.14.0：MIT，https://github.com/cloudflare/kumo 。使用官方 standalone CSS 与组件入口。
- React / React DOM 19.3.0：MIT，https://github.com/facebook/react 。
- Phosphor Icons React：MIT，https://github.com/phosphor-icons/react 。版本以 package-lock.json 为准。

旧版备份使用 Lucide；当前 Kumo 版已改为 Phosphor 通用图标。依赖许可副本在 `frontend/public/licenses/`。

- Wails v3 Go 模块 v3.0.0-alpha.63 / @wailsio/runtime 3.0.0-beta.28：MIT，https://github.com/wailsapp/wails 。版本分别以 go.mod 与 frontend/package-lock.json 为准；当前版本号不同，升级时核对原生桥接兼容性。

## Go 配置编辑与写入新增依赖

- Magpie JSON/JSONC 编辑算法：MIT，来源提交 a75fa9e54a7ec73a737681a3c345ae6fd0fe7d1f 的 internal/edit/json.go。改编到 internal/configedit/json.go，仅保留纯内存补丁函数，删除上游文件读取/写入入口；许可副本在 internal/configedit/LICENSE。没有引入上游代理或账号系统。
- tidwall/gjson 1.19.0 与 tidwall/jsonc 0.3.3：MIT，分别用于 JSON 路径读取、保留偏移的 JSONC 解析；依赖的 match/pretty 同属 MIT。来源：https://github.com/tidwall/gjson 、https://github.com/tidwall/jsonc 。
- gofrs/flock 0.13.0：BSD-3-Clause，用于本机 Connect 实例之间的文件写锁。来源：https://github.com/gofrs/flock 。
- go-keychain：MIT，https://github.com/keybase/go-keychain，macOS Security.framework 登录会话存储。
- go-keyring：MIT，https://github.com/zalando/go-keyring，非 macOS 系统凭据库接口。

- howett.net/plist 1.0.1：BSD-2-Clause，XML/二进制应用版本元数据解析；许可副本 internal/discovery/licenses/plist.txt。
