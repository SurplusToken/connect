package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/darwin/AppIcon.png
var appIcon []byte

func main() {
	app := application.New(application.Options{
		Name:        "SurplusToken Connect",
		Icon:        appIcon,
		Description: "连接你的 AI 编程工具",
		Assets:      application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		Mac:         application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
	})
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "SurplusToken Connect",
		Width: 880, Height: 640, MinWidth: 820, MinHeight: 600,
		BackgroundColour: application.NewRGB(255, 255, 255),
		URL:              "/",
	})
	// UI preview only: no gateway calls, client launches or credential writes.
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
