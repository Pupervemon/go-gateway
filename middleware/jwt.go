package middleware

import (
	"go-gateway/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 自定义 Token 载荷
// 需确保这里的字段名和 Java 端生成的 Token 载荷一致
type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"user_role"`
	jwt.RegisteredClaims
}

// JWTAuth 鉴权中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 【白名单检查】
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
			// 使用 Nacos 配置中的 SecretKey
			return []byte(config.AppConfig.Swust.Auth.SecretKey), nil
		})

		// 4. 校验 Token 有效性
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "Token 无效或已过期"})
			return
		}

		// 5. 存入 Context (透传给下游 Proxy 使用)
		// 在 proxy.go 中可以通过 c.Get("x-user-id") 取出并放入 Header
		c.Set("x-user-id", claims.UserID)
		c.Set("x-user-role", claims.Role)

		c.Next()
	}
}

// extractToken 从 Header 中提取 Bearer Token
func extractToken(c *gin.Context) string {
	bearToken := c.GetHeader("Authorization")
	// 通常 Authorization 头可能是 "Bearer <token>" 或者是直接 "<token>"
	if bearToken == "" {
		return ""
	}

	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 && strArr[0] == "Bearer" {
		return strArr[1]
	}
	return bearToken // 兼容没有 Bearer 前缀的情况
}

// isWhitelisted 辅助函数：检查路径是否在白名单中
func isWhitelisted(path string) bool {
	// 获取 Nacos 里配置的白名单列表
	whitelist := config.AppConfig.Swust.Auth.ExcludePaths

	for _, p := range whitelist {
		// 情况 1: 处理带参数的路径 (例如: /api/user/{userId}/profile)
		// 逻辑：截取 "{" 之前的部分作为前缀进行匹配
		// 效果：/api/user/ 匹配 /api/user/1001/profile
		if idx := strings.Index(p, "{"); idx != -1 {
			prefix := p[:idx]
			if strings.HasPrefix(path, prefix) {
				return true
			}
			continue
		}

		// 情况 2: 处理通配符 (例如: /api/rankings/**)
		// 逻辑：去掉末尾的 **，然后做前缀匹配
		cleanPrefix := strings.TrimSuffix(p, "**")
		if strings.HasPrefix(path, cleanPrefix) {
			return true
		}

		// 情况 3: 精确匹配 (例如: /api/user/login)
		// 上面的 HasPrefix 已经涵盖了精确匹配的情况，所以不需要额外写 equal 判断
	}
	return false
}
