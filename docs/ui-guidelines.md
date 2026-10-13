# 界面规范与参考

当前设计参考 ChatGPT 设置页的浅灰侧栏、白色内容区、细分隔线和圆角分组，采用 shadcn 风格的中性色。当前采用 React + Kumo 2.14.0，Select、分段 Tabs、Checkbox、Button、Input、Dialog 均使用组件库。通用图标采用 Phosphor，品牌图标保留原色。

## 本项目取值

- 默认窗口 880 × 640，最小 820 × 600；在保留侧栏和客户端三列布局的前提下减少横向留白。
- 遵循 Kumo design skill：正文、客户端名称和辅助文字 14px，标题 16–18px；优先系统字体。
- 背景白色，侧栏 #fafafa，正文 #27272a，辅助文字 #62626b。图标保留品牌原色，不将品牌标识全部转为灰度。
- 常用控件高度 32–36px，复选框 24px；保留键盘焦点轮廓。
- 弹窗限制键盘焦点在弹窗内，背景 inert，Escape 关闭并恢复触发按钮焦点。

## 规范依据

[Windows 字体规范](https://learn.microsoft.com/en-us/windows/apps/design/signature-experiences/typography)提供正文 14、辅助文字 12 的有效像素参考。[Apple 无障碍指南](https://developer.apple.com/design/human-interface-guidelines/accessibility)强调清晰可读、缩放和辅助技术支持。原生平台 pt/epx 与 CSS px 不应机械等同。

[WCAG 普通文字对比度](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html)要求一般文字至少 4.5:1，大字至少 3:1；它没有统一最低字号。[非文字对比度](https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html)对必要控件视觉信息要求 3:1。[最小点击目标](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html)为 24 × 24 CSS px，存在间距等例外；44 × 44 不是所有桌面控件的 AA 强制值。[文字缩放](https://www.w3.org/WAI/WCAG22/Understanding/resize-text.html)要求放大 200% 时不丢失内容或功能。

[shadcn 主题规范](https://ui.shadcn.com/docs/theming)采用背景、前景、主色、边框、弱化文字等语义色；借鉴其组织方式，不代表必须单色化品牌图标。

## 当前验证范围

Kumo 版已检查 1000px 与 589px 宽视口、模式切换前后的坐标、菜单选择、键盘打开、Escape 关闭与预览焦点恢复；通过 TypeScript、Vite 与 macOS arm64 构建。尚未完整验证屏幕阅读器、200% 缩放或旧版 WebView；不声明完整 WCAG 合规。

## Kumo 调研方式

读取官方 [kumo-design skill](https://github.com/cloudflare/kumo/blob/main/skills/kumo-design/SKILL.md)，运行 `npx @cloudflare/kumo@2.14.0 help`、`ai`、`doc Select`、`doc Tabs`、`doc Dialog`、`doc Checkbox`、`doc Button` 查 API。正文 14px、避免过重字重、相关文字紧凑排布、同心圆角及对话框保持挂载等均作为设计依据。

使用官方 standalone CSS 和组件独立入口；没有额外配置 Tailwind。后续可对比按需 CSS 构建，但体积报告记录当前实际产物，不使用预测优化数字。

### 桌面窗口滚动边界

根节点 html/body/#app 固定 100% 高度，overflow:hidden，overscroll-behavior:none。
workspace 使用固定高度的纵向 flex，标题栏不收缩；只有 main 使用 min-height:0、overflow-y:auto、overscroll-behavior:none。
不要让文档根节点承担正文滚动，否则 macOS WebView 回弹可能露出边缘，标题栏也会随文档移动。
参考：https://wails.io/docs/next/guides/overscroll/

已在 macOS 原生 Connect 窗口中加入反馈内容产生溢出，并检查向上、向下多页滚动：正文可滚动，标题栏和侧栏保持原位。
