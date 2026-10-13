# SuSu 品牌资源

`susu.png` 是用户提供的 SurplusToken 人物形象原图，1254 × 1254 PNG，原始内容保持不变。不要视为第三方开源图标或由依赖许可证覆盖。

`susu-warm.png` 为暖白背景版本，保留原始贴边裁切构图；保留作对比。深色原图仍保留在 `susu.png`。

`susu-white.png` 为当前 Dock / App 图标使用的中性近白背景版本，参考日历、Safari 等浅底图标；侧栏 32px 头像也使用同一近白背景版本。

运行 `make icons`（macOS，需要系统自带 sips 与 iconutil）生成：

- `frontend/public/brand/susu.png`：128 × 128，用于应用侧栏。
- `frontend/public/brand/favicon.png`：32 × 32，用于浏览器预览。
- `build/darwin/AppIcon.icns`：16、32、128、256、512 点对应的 1x/2x 图层，用于 Finder / Dock。

`make package-mac` 会重新生成并复制 ICNS 到应用包 Resources，Info.plist 用 CFBundleIconFile 引用。App 图标在 1024px 透明画布中使用 824px 图像区域和圆角蒙版，避免 Dock 中满幅方形过大；不重绘人物。运行时与 ICNS 共用此图，防止 Wails 覆盖成原始方图。Windows EXE 图标资源尚未接入 Windows 构建流程。
