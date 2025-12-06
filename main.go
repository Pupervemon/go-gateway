package main

import (
	"fmt"
	"go-gateway/config"
	"go-gateway/pkg/nacos"
	"go-gateway/router"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	config.InitConfig()

	// 2. 初始化 Nacos
	nacos.InitNacos()

	// 3. 初始化 Web 引擎
	r := gin.Default() // 默认带有 Logger 和 Recovery 中间件

	// 4. 注册路由
	router.InitRouter(r)

	// 5. 启动服务
	port := config.AppConfig.Server.Port
	fmt.Printf("🚀 Go Gateway running on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		panic(err)
	}
}
