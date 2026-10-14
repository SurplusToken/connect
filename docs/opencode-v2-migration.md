# OpenCode v2 升级与接入

2026-10-13：本机 npm 安装从 opencode-ai 1.16.2 替换为 @opencode/cli 2.0.28；实际 opencode --version 返回 opencode v2.0.28。

依照[官方迁移说明](https://opencode.ai/v2/docs/migrate-v1/)先卸载 v1 包再安装 v2 包；两个版本共用命令和用户配置位置。升级前原配置已备份至 ~/Library/Application Support/SurplusToken Connect/upgrades/opencode-v1-20261013-204409，目录 0700、文件 0600。manifest.json 记录来源、备份与回退命令。本轮未转换或覆盖真实用户配置，未启动真实工作会话。

Connect 当前仅开放 v2 配置。按[原生 provider 文档](https://opencode.ai/v2/docs/providers/)写入 providers.surplustoken.package/settings、模型目录；默认模型和标题模型不写入模板；保留其他 provider、agent、注释与用户字段。OpenAI 分组使用 Responses 包，Anthropic 使用 Messages 兼容包，DeepSeek/Kimi 使用 Chat 兼容包。混合分组暂不开放，避免猜测协议。

客户端列表改为每个客户端独立选择 API Key；兼容 Key 根据分组声明筛选，Key 明文不进入前端。OpenCode 只需选 Key，再应用供应商模板；模型由客户端的 /models 选择。不同客户端可以使用不同 Key。

验证：隔离目录中给 Claude Code 与 OpenCode 分别使用两个不同 Key，预览、写入和字段保留通过；OpenCode 恢复后原件字节完全相同。安装的 v2 CLI 在隔离 HOME/XDG 目录执行 models --standalone，成功加载原生 provider 并列出 surplustoken/gpt-fixture。该夹具端点指向不可用的本机端口，无真实凭据，无推理请求。

此次验证不代表所有项目覆盖、插件或推理路由均已验收。v1 插件需要另行迁移到 v2 API；模型清单不是协议兼容性的证明。
