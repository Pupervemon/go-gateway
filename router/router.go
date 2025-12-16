package router

import (
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
// 使用动态路由配置，从Nacos加载路由规则
func InitRouter(r *gin.Engine) {
	// 使用动态路由加载器
	InitDynamicRouter(r)
}
