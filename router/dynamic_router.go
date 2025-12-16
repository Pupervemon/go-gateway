package router

import (
	"go-gateway/config"
	"go-gateway/middleware"
	"go-gateway/proxy"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// InitDynamicRouter 初始化动态路由
func InitDynamicRouter(r *gin.Engine) {
	// 统一入口：全部以 /api 开头
	api := r.Group("/api")

	// 挂载JWT中间件
	api.Use(middleware.JWTAuth())

	// 获取路由配置
	routesConfig := config.AppConfig.Routes

	// 按order字段排序，order小的优先
	sortRoutesByOrder(routesConfig.Routes)

	// 遍历路由配置并注册路由
	for _, route := range routesConfig.Routes {
		registerRoute(api, route)
	}
}

// sortRoutesByOrder 按order字段排序路由
func sortRoutesByOrder(routes []config.RouteConfig) {
	sort.Slice(routes, func(i, j int) bool {
		// 如果 order 相同，按 ID 排序保证稳定
		if routes[i].Order == routes[j].Order {
			return routes[i].ID < routes[j].ID
		}
		return routes[i].Order < routes[j].Order
	})
}

// registerRoute 注册单个路由
func registerRoute(api *gin.RouterGroup, route config.RouteConfig) {
	// 解析URI
	targetURL, serviceName, isDirectURL := parseURI(route.URI)

	// 遍历所有Path断言
	for _, predicate := range route.Predicates {
		if predicate.Name == "Path" {
			// 获取所有路径参数
			for _, arg := range predicate.Args {
				if path, ok := arg.(string); ok {
					// 创建路由处理器
					handler := createRouteHandler(targetURL, serviceName, isDirectURL, route.Filters)

					// 将路径转换为Gin路由格式
					ginPath := convertPathToGinFormat(path)

					// 注册路由
					api.Any(ginPath, handler)
				}
			}
		}
	}
}

// parseURI 解析URI，区分直接URL和服务发现
func parseURI(uri string) (targetURL string, serviceName string, isDirectURL bool) {
	// 检查是否是lb://开头的负载均衡URL
	if strings.HasPrefix(uri, "lb://") {
		// 提取服务名，如 lb://user-service -> user-service
		serviceName = strings.TrimPrefix(uri, "lb://")
		return "", serviceName, false
	}

	// 直接URL（如 https://api.dify.ai/v1）
	return uri, "", true
}

// createRouteHandler 创建路由处理器
func createRouteHandler(targetURL string, serviceName string, isDirectURL bool, filters []config.Filter) gin.HandlerFunc {
	// 如果是直接URL，使用新的代理处理器
	if isDirectURL {
		return proxy.NewDirectProxyHandler(targetURL)
	}

	// 否则使用原有的服务发现处理器
	return proxy.NewProxyHandler(serviceName, "")
}

// convertPathToGinFormat 将路径格式转换为Gin路由格式
func convertPathToGinFormat(path string) string {
	// 移除开头的/api/（因为已经在api组里）
	path = strings.TrimPrefix(path, "/api/")

	// 确保路径以 / 开头（Gin路由要求）
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 处理通配符路径
	if strings.HasSuffix(path, "/**") {
		// 将 Spring Cloud Gateway 的 /** 转换为 Gin 的 /*path
		path = strings.TrimSuffix(path, "/**")
		if path == "" {
			path = "/*path"
		} else {
			path = path + "/*path"
		}
		return path
	}

	// 对于没有通配符的路径，我们返回精确路径
	// 让 Gin 路由树自动处理子路径的匹配
	return path
}

