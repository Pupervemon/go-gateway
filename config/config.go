package config

import (
	"fmt"
	"log"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// AppConfig 全局配置单例
// 其他包使用时直接调用 config.AppConfig.Swust.Auth.xxx
var AppConfig Config

// Config 总配置结构体
type Config struct {
	// --- 本地配置 (Bootstrapping) ---
	Server ServerConfig `mapstructure:"server" yaml:"server"`
	Nacos  NacosConfig  `mapstructure:"nacos" yaml:"nacos"`

	// --- 远程业务配置 (From Nacos gateway.yaml) ---
	// 这里对应你截图中的 swust: 结构
	Swust SwustConfig `mapstructure:"swust" yaml:"swust"`
}

type ServerConfig struct {
	Port string `mapstructure:"port" yaml:"port"`
}

type NacosConfig struct {
	Host      string `mapstructure:"host" yaml:"host"`
	Port      uint64 `mapstructure:"port" yaml:"port"`
	Namespace string `mapstructure:"namespace" yaml:"namespace"`
	Group     string `mapstructure:"group" yaml:"group"`
	DataId    string `mapstructure:"dataid" yaml:"dataid"` // 对应 gateway.yaml
}

type SwustConfig struct {
	Auth AuthConfig `mapstructure:"auth" yaml:"auth"`
}

type AuthConfig struct {
	// 注意 tag 要和你 Nacos 里的 key 完全一致
	DifyApiKey              string   `mapstructure:"dify-api-key" yaml:"dify-api-key"`
	SecretKey               string   `mapstructure:"secret-key" yaml:"secret-key"`
	AccessTokenExpirationMs int64    `mapstructure:"access-token-expiration-ms" yaml:"access-token-expiration-ms"`
	ExcludePaths            []string `mapstructure:"exclude-paths" yaml:"exclude-paths"`
}

// InitConfig 初始化配置流程
func InitConfig() {
	// -------------------------------------------------------
	// 第一步：使用 Viper 读取本地 config.yaml (获取 Nacos 地址)
	// -------------------------------------------------------
	viper.SetConfigName("config") // 文件名
	viper.SetConfigType("yaml")   // 类型
	viper.AddConfigPath("config") // 路径 config/

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("❌ 读取本地配置文件失败: %v", err)
	}

	// 将本地配置解析到 AppConfig 中
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("❌ 解析本地配置失败: %v", err)
	}

	fmt.Printf("✅ [Local] 本地配置加载完成，Nacos地址: %s:%d\n", AppConfig.Nacos.Host, AppConfig.Nacos.Port)

	// -------------------------------------------------------
	// 第二步：根据本地配置连接 Nacos，读取远程业务配置
	// -------------------------------------------------------
	loadFromNacos()
}

func loadFromNacos() {
	// 1. 构建 Nacos Server 配置
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(AppConfig.Nacos.Host, AppConfig.Nacos.Port, constant.WithContextPath("/nacos")),
	}

	// 2. 构建 Nacos Client 配置
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(AppConfig.Nacos.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("./logs/nacos/log"),
		constant.WithCacheDir("./logs/nacos/cache"),
		constant.WithLogLevel("error"), // 设置为 error 减少控制台噪音
	)

	// 3. 创建 Config Client (配置中心客户端)
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		log.Fatalf("❌ 创建 Nacos Config Client 失败: %v", err)
	}

	// 4. 拉取配置
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: AppConfig.Nacos.DataId,
		Group:  AppConfig.Nacos.Group,
	})
	if err != nil {
		log.Fatalf("❌ 从 Nacos 拉取配置失败 (DataId: %s): %v", AppConfig.Nacos.DataId, err)
	}

	// 5. 【核心】解析 Nacos 返回的 YAML 字符串，合并到 AppConfig
	// 使用 yaml.Unmarshal 直接将远程配置覆盖/填充到 AppConfig 指针中
	err = yaml.Unmarshal([]byte(content), &AppConfig)
	if err != nil {
		log.Fatalf("❌ 解析 Nacos YAML 内容失败: %v", err)
	}

	fmt.Println("✅ [Remote] Nacos 配置加载成功！")
	fmt.Printf("   >> Auth Secret: %s...\n", AppConfig.Swust.Auth.SecretKey[0:5]) // 打印前几位验证
	fmt.Printf("   >> 白名单路径数量: %d\n", len(AppConfig.Swust.Auth.ExcludePaths))

	// 6. (可选) 监听配置变化实现热更新
	err = client.ListenConfig(vo.ConfigParam{
		DataId: AppConfig.Nacos.DataId,
		Group:  AppConfig.Nacos.Group,
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Println("⚠️  监听到 Nacos 配置发生变化，正在刷新...")
			// 重新解析最新配置
			if err := yaml.Unmarshal([]byte(data), &AppConfig); err == nil {
				fmt.Println("✅  配置刷新成功！")
			} else {
				fmt.Printf("❌  配置刷新失败: %v\n", err)
			}
		},
	})
}
