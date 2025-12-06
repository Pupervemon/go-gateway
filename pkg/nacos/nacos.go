package nacos

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var NamingClient naming_client.INamingClient

// InitNacos 初始化 Nacos 客户端
func InitNacos() {
	// 创建 ServerConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: "127.0.0.1",
			Port:   8848,
		},
	}

	// 创建 ClientConfig
	clientConfig := constant.ClientConfig{
		NamespaceId:         "", // 如果需要可配置命名空间ID
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "./logs/nacos",
		CacheDir:            "./cache/nacos",
		LogLevel:            "info",
	}

	// 创建命名服务客户端
	var err error
	NamingClient, err = clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)

	if err != nil {
		panic(fmt.Sprintf("Failed to create Nacos client: %v", err))
	}

	fmt.Println("✅ Nacos client initialized successfully")
}

// GetInstance 从 Nacos 获取服务实例
func GetInstance(serviceName string) (string, uint64, error) {
	instance, err := NamingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
	})

	if err != nil {
		return "", 0, fmt.Errorf("failed to get instance for service %s: %v", serviceName, err)
	}

	return instance.Ip, instance.Port, nil
}
