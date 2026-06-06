package main

import (
	"log"
	"net/http"
	"os"

	"github.com/athena/staticman/internal/auth"
	"github.com/athena/staticman/internal/config"
	"github.com/athena/staticman/internal/handler"
	"github.com/athena/staticman/internal/middleware"
	"github.com/athena/staticman/internal/web"
)

func main() {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}

	cfg, err := config.Load(dataDir)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	accessKey := os.Getenv("ACCESS_KEY")
	if accessKey == "" {
		accessKey = cfg.AccessKeyHash
	}

	authSvc := auth.New(cfg, accessKey)

	// 启动配置热加载
	cfg.Watch()

	mux := http.NewServeMux()

	// 注册所有路由（API、原始文件层、兼容层）
	h := handler.New(cfg, authSvc)
	h.RegisterRoutes(mux)

	// 前端 SPA 静态资源
	spaHandler := web.NewSPAHandler()
	mux.Handle("/", spaHandler)

	// 应用中间件（不再需要 Auth 中间件，由各 handler 自行处理）
	var finalHandler http.Handler = mux
	finalHandler = middleware.CORS(finalHandler)
	finalHandler = middleware.Logging(finalHandler)

	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":8080"
	}
	log.Printf("StaticMan v2 starting on %s", addr)
	log.Printf("Routes: /api/* (Web UI with auth), /{category}/* (raw files, ?key= for protected), /d/* (legacy compat)")
	if err := http.ListenAndServe(addr, finalHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}