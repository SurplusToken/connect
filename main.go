package main

import (
	"embed"
	"log"
	"surplustoken-connect/internal/configuration"
	"surplustoken-connect/internal/gateway"
	"surplustoken-connect/internal/httpapi"
	"surplustoken-connect/internal/sessionstore"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/darwin/AppIcon.png
var appIcon []byte

func main() {
	engine, err := configuration.Default()
	if err != nil {
		log.Fatal("无法初始化本地配置服务")
	}
	api := httpapi.New(gateway.New(), engine)
	api.Store = sessionstore.System{}
	app := application.New(application.Options{
		Name:        "SurplusToken Connect",
		Icon:        appIcon,
		Description: "连接你的 AI 编程工具",
		Assets:      application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		Services:    []application.Service{application.NewService(httpapi.NewBridge(api))},
		Mac:         application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
	})
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "SurplusToken Connect",
		Width: 880, Height: 640, MinWidth: 820, MinHeight: 600,
		BackgroundColour: application.NewRGB(255, 255, 255),
		URL:              "/",
	})
	// All writes require a server-issued preview plan; Codex remains protected.
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
