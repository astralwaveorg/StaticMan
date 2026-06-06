package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/athena/staticman/internal/auth"
	"github.com/athena/staticman/internal/cache"
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

	// 创建内存缓存
	appCache := cache.New()

	mux := http.NewServeMux()

	// 注册所有路由（API、原始文件层、兼容层）
	h := handler.New(cfg, authSvc, appCache)
	h.RegisterRoutes(mux)

	// 前端 SPA 静态资源（动态注入站点标题和 logo）
	spaHandler := web.NewSPAHandler(func() web.SiteConfig {
		site := cfg.GetSite()
		return web.SiteConfig{
			TitleCN:     site.TitleCN,
			TitleEN:     site.TitleEN,
			Title:       site.Title,
			Description: site.Description,
			Logo:        site.Logo,
		}
	})
	mux.Handle("/", spaHandler)

	// 限流配置
	rateLimits := map[string]*middleware.RateLimiter{
		"/api/auth":  middleware.NewRateLimiter(5, 5.0/60.0),   // 5次/分钟
		"/api/search": middleware.NewRateLimiter(30, 30.0/60.0), // 30次/分钟
		"/api/":      middleware.NewRateLimiter(120, 120.0/60.0), // 120次/分钟
	}

	// 应用中间件
	var finalHandler http.Handler = mux
	finalHandler = middleware.RateLimitMiddleware(rateLimits)(finalHandler)
	finalHandler = middleware.CORS(finalHandler)
	finalHandler = middleware.Logging(finalHandler)

	// 定期清理限流器
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			for _, limiter := range rateLimits {
				limiter.Cleanup()
			}
		}
	}()

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