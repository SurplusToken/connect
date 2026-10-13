# Kumo 设计审查

2026-10-13，针对当前 React + Kumo 2.14.0 界面。依据官方 [kumo-design/SKILL.md](https://github.com/cloudflare/kumo/blob/main/skills/kumo-design/SKILL.md)，结合源码与浏览器实际计算样式检查。使用组件库并不自动意味着每处符合规范。

| 规则 | 当前检查与处理 |
| --- | --- |
| 内容文字 14px | 正文、按钮、输入框、标签、模型选择 14px；标题 16/18px。浏览器核查按钮和表单实际样式，不只看 CSS 声明。 |
| 标题 sentence case | 中文标题自然书写，产品名 SurplusToken Connect；无全大写装饰标题。 |
| 不修改字距 | 自有样式未设置 letter-spacing / tracking。 |
| 不用粗黑字重 | 标题 600，强调与控件 500，正文 400；无自有 font-bold / 700。 |
| 相关文字更紧凑 | 空状态标题和说明改为独立文字组、间距 6px，与图标和按钮保持更大距离。 |
| 留白考虑行高 | 面板改 16px/20px、弹窗 20px/24px、预览项和提示 10px/14px（纵/横）；不机械四边等宽。空状态的大留白用于区域布局。 |
| 悬停不做颜色渐变 | 按钮计算 transition-duration 为 0s；保留组件库的选中指示器运动，不额外添加颜色渐变。 |
| 阴影不用实体边框叠加 | 自有面板仅细边框、无投影；浮层由 Kumo 默认 ring/shadow 管理。 |
| 同心圆角 | 列表外框 10px，内行背景 9px，间隔为 1px 边框；未覆盖 Tabs 内部圆角。 |
| 图标对齐首行 | 可能换行的 Codex 提示改为首行 line-height 高度的图标容器，图标 14px；品牌图片为独立身份图标，不按行内文字图标缩小。 |
| 行内等宽字略小 | 当前配置路径为独立代码块，修正为 14px；只有混排行内代码才适用约 0.9em。 |
| 固定区域分隔 | 固定侧栏有边界；页头有底部分隔线。 |
| 折叠过程保持尺寸 | 当前没有宽度折叠动画；响应式侧栏按断点切换。 |
| 不叠套 LayerCard | 没有 LayerCard 嵌套。 |
| Dialog 不条件挂载 | Dialog.Root 持续挂载，以 open/onOpenChange 控制；没有 open && Dialog。 |

## 验证边界

本轮完成 TypeScript/Vite 与 macOS arm64 构建、Info.plist 校验及 SuSu 资源生成检查。前轮已检查选择器键盘操作、模式切换坐标、Escape 和焦点恢复。设计规则审查不等于完整无障碍认证，尚未覆盖屏幕阅读器、旧 WebView 和 Windows 实机。

旧版与首次 Kumo 体积对比快照仍保留在 `.references/comparison/`；本轮加入 SuSu 的 ICNS 后体积会增加，旧表不代表当前图标版。

## 侧栏补充

参考 Kumo Sidebar 的 active 状态和独立键盘焦点设计，保留紧凑布局。当前页面使用更清楚的灰底与左侧深色标记；focus-visible 使用 2px 描边，Tab 移动焦点不改变当前页面。浏览器已检查两种状态同时显示。侧栏小头像使用深底原图，Dock 使用中性近白版。
