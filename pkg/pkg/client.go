package nacos

import (
	"fmt"
	"go-gateway/config"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var NamingClient naming_client.INamingClient

func InitNacos() {
	cfg := config.AppConfig.Nacos

	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "tmp/nacos/log",
		CacheDir:            "tmp/nacos/cache",
		LogLevel:            "error",
	}

	serverConfigs := []constant.ServerConfig{
		{IpAddr: cfg.Host, Port: cfg.Port},
	}

	var err error
	NamingClient, err = clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Nacos 初始化失败: %v", err))
	}
}

// GetInstance 获取健康实例 (负载均衡)
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
