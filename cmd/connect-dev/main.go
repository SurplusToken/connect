// connect-dev serves only the local UI API; it never launches a client.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"surplustoken-connect/internal/configuration"
	"surplustoken-connect/internal/gateway"
	"surplustoken-connect/internal/httpapi"
	"time"
)

func main() {
	fixture := flag.String("fixture-dir", "", "配置写入测试目录（不使用真实客户端配置）")
	mock := flag.Bool("mock", false, "使用离线网关夹具（必须指定 fixture-dir）")
	flag.Parse()
	if *mock && *fixture == "" {
		log.Fatal("离线夹具必须指定 fixture-dir")
	}
	var engine *configuration.Engine
	var e error
	if *fixture != "" {
		_ = os.MkdirAll(*fixture, 0700)
		root, err := filepath.EvalSymlinks(*fixture)
		if err != nil {
			log.Fatal("测试目录不可用")
		}
		engine, e = configuration.New(filepath.Join(root, "backups"), map[string]string{"claude": filepath.Join(root, "claude", "settings.json"), "opencode": filepath.Join(root, "opencode", "opencode.jsonc")})
	} else {
		engine, e = configuration.Default()
	}
	if e != nil {
		log.Fatal("无法初始化本地配置服务")
	}
	g := gateway.New()
	if *mock {
		remote := mockGateway()
		defer remote.Close()
		g.Base = remote.URL
	}
	api := httpapi.New(g, engine)
	api.Fixture = *fixture != ""
	server := &http.Server{Addr: "127.0.0.1:43129", Handler: api, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 30 * time.Second}
	log.Print("Connect API: http://127.0.0.1:43129（仅本机）")
	log.Fatal(server.ListenAndServe())
}
