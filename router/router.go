package router

import (
	"go-gateway/middleware"
	"go-gateway/proxy"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	// 注册全局中间件 (比如 CORS, Logger)
	// r.Use(middleware.Cors())

	// ==========================================
	// 1. 需要鉴权的 API 组 (需要 Header 带 Token)
	// ==========================================
	api := r.Group("/api")
	api.Use(middleware.JWTAuth()) // 挂载 JWT 中间件
	{
		// 【关键修改】：
		// Java 端配置了 server.servlet.context-path: /api
		// 所以 Java 真实的监听地址是 /api/user/xxx
		// 网关收到的请求也是 /api/user/xxx
		// 结论：不需要剥离前缀，原样转发即可（第二个参数传 ""）

		// 路由: /api/user/** -> user-service /api/user/**
		api.Any("/user/*path", proxy.NewProxyHandler("user-service", ""))

		// 路由: /api/problem/** -> problem-service /api/problem/**
		api.Any("/problem/*path", proxy.NewProxyHandler("problem-service", ""))

		// 路由: /api/judge/** -> judge-service /api/judge/**
		api.Any("/judge/*path", proxy.NewProxyHandler("judge-service", ""))
	}

	// ==========================================
	// 2. 公开接口 (如登录注册，不需要 Token)
	// ==========================================
	auth := r.Group("/auth")
	{
		// 注意：这里的转发逻辑取决于你希望前端怎么请求
		// 假设前端请求 /auth/user/login
		// 我们希望转发给 Java 的是 /api/user/login (因为 Java 只认 /api 开头)

		// 这里的处理可能需要你的 ProxyHandler 支持“重写路径”功能
		// 如果目前的 ProxyHandler 只是简单的 stripPrefix，
		// 那么直接转发会导致发过去 /auth/user/login (Java 会报 404)

		// 临时方案：如果你的 Proxy 只支持 Strip，
		// 你可能需要让前端直接请求 /api/user/login 并在 JWTAuth 中排除登录接口
		// 或者修改 ProxyHandler 支持把 /auth 替换为 /api

		// 暂时保持原样，先把 404 搞定
		auth.Any("/*path", proxy.NewProxyHandler("user-service", ""))
	}
}
