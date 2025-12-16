package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// 全局配置变量
var AppConfig *Config

// Config 总配置结构体
type Config struct {
	Server ServerConfig `mapstructure:"server" json:"server" yaml:"server"`
	Nacos  NacosConfig  `mapstructure:"nacos" json:"nacos" yaml:"nacos"`
	Swust  SwustConfig  `mapstructure:"swust" json:"swust" yaml:"swust"`
	Routes RoutesConfig `mapstructure:"routes" json:"routes" yaml:"routes"`
}

// ServerConfig 服务器端口配置
type ServerConfig struct {
	Port string `mapstructure:"port" json:"port" yaml:"port"`
}

// NacosConfig Nacos 连接配置
type NacosConfig struct {
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Port         uint64 `mapstructure:"port" json:"port" yaml:"port"`
	Namespace    string `mapstructure:"namespace" json:"namespace" yaml:"namespace"`
	Group        string `mapstructure:"group" json:"group" yaml:"group"`
	DataId       string `mapstructure:"dataid" json:"dataId" yaml:"dataid"`
	RoutesDataId string `mapstructure:"routes-dataid" json:"routesDataId" yaml:"routes-dataid"`
}

// SwustConfig 业务配置
type SwustConfig struct {
	Auth AuthProperties `mapstructure:"auth" json:"auth" yaml:"auth"`
}

// AuthProperties 认证相关配置
type AuthProperties struct {
	IncludePaths   []string `mapstructure:"include-paths" json:"includePaths" yaml:"include-paths"`
	ExcludePaths   []string `mapstructure:"exclude-paths" json:"excludePaths" yaml:"exclude-paths"`
	SecretKey      string   `mapstructure:"secret-key" json:"secretKey" yaml:"secret-key"`
	DifyApiKey     string   `mapstructure:"dify-api-key" json:"difyApiKey" yaml:"dify-api-key"`
	DifyPathPrefix string   `mapstructure:"dify-path-prefix" json:"difyPathPrefix" yaml:"dify-path-prefix"`
}

// InitConfig 初始化配置
func InitConfig() {
	v := viper.New()
	v.AddConfigPath("./config") // 配置文件路径
	v.SetConfigName("config")   // 配置文件名 (不带后缀)
	v.SetConfigType("yaml")     // 配置文件类型

	// 读取本地 config.yaml
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("❌ 读取本地配置失败: %v", err)
	}

	// 绑定环境变量
	v.AutomaticEnv()

	// 解析到结构体
	if err := v.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("❌ 配置解析失败: %v", err)
	}

	// 设置默认值
	if AppConfig.Swust.Auth.DifyPathPrefix == "" {
		AppConfig.Swust.Auth.DifyPathPrefix = "/api/dify/"
	}

	// 设置路由配置默认值
	if AppConfig.Nacos.RoutesDataId == "" {
		AppConfig.Nacos.RoutesDataId = "routes-go.json"
	}

	fmt.Printf("✅ 本地配置加载成功，Nacos地址: %s:%d\n", AppConfig.Nacos.Host, AppConfig.Nacos.Port)
}
