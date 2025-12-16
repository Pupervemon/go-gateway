package nacos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-gateway/config"
	"log"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper" // 用来解析拉取到的 YAML
)

var (
	NamingClient naming_client.INamingClient
	ConfigClient config_client.IConfigClient // 配置中心客户端
)

func InitNacos() {
	cfg := config.AppConfig.Nacos

	// 1. 公共配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "logs/nacos",
		CacheDir:            "logs/nacos/cache",
		LogLevel:            "error",
	}

	serverConfigs := []constant.ServerConfig{
		{IpAddr: cfg.Host, Port: cfg.Port},
	}

	var err error

	// 2. 初始化服务发现
	NamingClient, err = clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Nacos NamingClient 初始化失败: %v", err))
	}

	// 3. 初始化配置中心
	ConfigClient, err = clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Nacos ConfigClient 初始化失败: %v", err))
	}

	// 拉取远程配置并生效
	loadRemoteConfig(cfg.DataId, cfg.Group)

	// 拉取路由配置并生效
	loadRouteConfig(cfg.RoutesDataId, cfg.Group)
}

// 拉取并解析远程配置
func loadRemoteConfig(dataId, group string) {
	// A. 拉取字符串内容
	content, err := ConfigClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		panic(fmt.Sprintf("拉取 Nacos 配置失败 DataId=%s: %v", dataId, err))
	}

	if content == "" {
		log.Printf("⚠️ 警告: Nacos 远程配置为空 (DataId: %s)", dataId)
		return
	}

	fmt.Printf("🔥 成功拉取远程配置 DataId: %s\n", dataId)

	// B. 使用 Viper 解析 YAML 字符串
	v := viper.New()
	v.SetConfigType("yaml") // 极其重要：告诉 Viper 内容是 YAML

	if err := v.ReadConfig(bytes.NewBufferString(content)); err != nil {
		panic(fmt.Sprintf("远程配置解析失败: %v", err))
	}

	// C. 合并到全局变量
	// 这步会将 Nacos 里的 swust.auth 配置注入到 config.AppConfig
	if err := v.Unmarshal(&config.AppConfig); err != nil {
		panic(fmt.Sprintf("配置合并失败: %v", err))
	}

	// D. 开启动态监听 (热更新)
	go listenConfigChange(dataId, group)
}

// 监听配置变化
func listenConfigChange(dataId, group string) {
	err := ConfigClient.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Println("⚠️ Nacos 配置发生变更，正在刷新...")

			// 重新解析逻辑
			v := viper.New()
			v.SetConfigType("yaml")
			v.ReadConfig(bytes.NewBufferString(data))

			// 更新内存中的配置
			// 注意：生产环境这里可能需要加锁
			v.Unmarshal(&config.AppConfig)

			fmt.Println("✅ 配置热更新完成")
		},
	})
	if err != nil {
		log.Printf("监听配置失败: %v", err)
	}
}

// 拉取并解析路由配置
func loadRouteConfig(dataId, group string) {
	// A. 拉取字符串内容
	content, err := ConfigClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		panic(fmt.Sprintf("拉取 Nacos 路由配置失败 DataId=%s: %v", dataId, err))
	}

	if content == "" {
		log.Printf("⚠️ 警告: Nacos 路由配置为空 (DataId: %s)", dataId)
		return
	}

	fmt.Printf("🔥 成功拉取路由配置 DataId: %s\n", dataId)

	// B. 直接使用 json.Unmarshal 解析 JSON 数组
	var routes []config.RouteConfig
	if err := json.Unmarshal([]byte(content), &routes); err != nil {
		panic(fmt.Sprintf("路由配置解析失败: %v\n内容: %s", err, content))
	}

	// C. 更新全局配置中的路由
	config.AppConfig.Routes = config.RoutesConfig{
		Routes: routes,
	}

	// D. 开启动态监听 (热更新)
	go listenRouteChange(dataId, group)
}

// 监听路由配置变化
func listenRouteChange(dataId, group string) {
	err := ConfigClient.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Println("⚠️ Nacos 路由配置发生变更，正在刷新...")

			// 直接使用 json.Unmarshal 解析
			var routes []config.RouteConfig
			if err := json.Unmarshal([]byte(data), &routes); err != nil {
				log.Printf("路由配置热更新失败: %v", err)
				return
			}

			// 更新内存中的路由配置
			config.AppConfig.Routes = config.RoutesConfig{
				Routes: routes,
			}

			fmt.Println("✅ 路由配置热更新完成")
		},
	})
	if err != nil {
		log.Printf("监听路由配置失败: %v", err)
	}
}

// GetInstance 获取健康实例
func GetInstance(serviceName string) (string, uint64, error) {
	instance, err := NamingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   config.AppConfig.Nacos.Group,
	})
	if err != nil {
		return "", 0, err
	}
	return instance.Ip, instance.Port, nil
}
