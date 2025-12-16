package proxy

import (
	"fmt"
	"go-gateway/pkg/nacos"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// NewProxyHandler 创建反向代理处理器
// targetService: 目标微服务名称
// stripPrefix: 需要移除的路径前缀 (如 /api/user -> /user)
func NewProxyHandler(targetService string, stripPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 服务发现
		ip, port, err := nacos.GetInstance(targetService)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "服务实例不可用"})
			return
		}

		// 2. 构建目标 URL
		targetStr := fmt.Sprintf("http://%s:%d", ip, port)
		targetUrl, _ := url.Parse(targetStr)

		// 3. 创建反向代理
		proxy := httputil.NewSingleHostReverseProxy(targetUrl)

		// =======================================================================
		// 自定义 Transport
		// 目的：强制不使用系统代理 (VPN)，防止 Go 将局域网 IP 发送给代理软件导致 502
		// =======================================================================
		transport := &http.Transport{
			Proxy: nil, // <--- 重点：设置为 nil，忽略 HTTP_PROXY 等环境变量
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second, // 建立连接的超时时间
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		proxy.Transport = transport

		// 4. 定制请求 (Director)
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)

			// A. 透传 Header
			if userId, exists := c.Get("x-user-id"); exists {
				req.Header.Set("user_id", fmt.Sprintf("%d", userId))
			}
			if role, exists := c.Get("x-user-role"); exists {
				req.Header.Set("x-user-role", role.(string))
			}

			// B. 路径重写 (StripPrefix)
			// 注意：如果你的 Java 服务配置了 context-path，请谨慎使用 StripPrefix
			if stripPrefix != "" {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, stripPrefix)
				req.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, stripPrefix)
			}
		}

		// 5. 错误处理
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			// 这里可以打印一下具体的错误日志，方便排查
			fmt.Printf("Proxy Error: %v\n", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "网关转发异常"})
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// NewDirectProxyHandler 创建直接URL代理处理器（用于外部服务如Dify）
func NewDirectProxyHandler(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 解析目标URL
		url, err := url.Parse(targetURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无效的目标URL"})
			return
		}

		// 2. 创建反向代理
		proxy := httputil.NewSingleHostReverseProxy(url)

		// 3. 配置Transport（不使用系统代理）
		transport := &http.Transport{
			Proxy: nil,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		proxy.Transport = transport

		// 4. 定制请求
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)

			// 透传用户信息（如果有）
			if userId, exists := c.Get("x-user-id"); exists {
				req.Header.Set("user_id", fmt.Sprintf("%d", userId))
			}
			if role, exists := c.Get("x-user-role"); exists {
				req.Header.Set("x-user-role", role.(string))
			}

			// 重写路径，移除/api前缀
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api")
			if req.URL.RawPath != "" {
				req.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, "/api")
			}
		}

		// 5. 错误处理
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			fmt.Printf("Direct Proxy Error: %v\n", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "网关转发异常"})
		}

		// 6. 处理请求
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
