package middleware

import (
	"go-gateway/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 【新增】白名单检查
		// 如果当前路径在 Nacos 配置的 exclude-paths 里，直接放行
		if isWhitelisted(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 2. 提取 Token
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "未授权: Token 缺失"})
			return
		}

		// 3. 解析 Token
		claims := &CustomClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			// 【修复点】这里要用新的配置路径
			// 之前是 config.AppConfig.Jwt.Secret
			// 现在是 config.AppConfig.Swust.Auth.SecretKey
			return []byte(config.AppConfig.Swust.Auth.SecretKey), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "Token 无效或已过期"})
			return
		}

		// 4. 存入 Context (透传给下游)
		c.Set("x-user-id", claims.UserID)
		c.Set("x-user-role", claims.Role)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	bearToken := c.GetHeader("Authorization")
	// 通常 Authorization 头可能是 "Bearer <token>" 或者是直接 "<token>"，做一个兼容处理
	if bearToken == "" {
		return ""
	}

	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 && strArr[0] == "Bearer" {
		return strArr[1]
	}
	return bearToken // 兼容没有 Bearer 前缀的情况
}

// 辅助函数：检查路径是否在白名单中
func isWhitelisted(path string) bool {
	// 获取 Nacos 里配置的白名单列表
	whitelist := config.AppConfig.Swust.Auth.ExcludePaths

	for _, p := range whitelist {
		// 简单的匹配逻辑：精确匹配 或 前缀匹配
		// 比如配置了 /api/user/login，那么请求 /api/user/login 就会匹配
		// 如果你想支持 /** 通配符，需要更复杂的正则匹配，这里先做简单的包含判断

		// 移除 Nacos 配置里可能存在的 ** 通配符以便做前缀匹配
		cleanPrefix := strings.TrimSuffix(p, "**")

		if strings.HasPrefix(path, cleanPrefix) {
			return true
		}
	}
	return false
}
