package config

import (
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Server *ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	ListenAddr string `mapstructure:"listen_addr"`
}

var c Config

func init() {
	c = initConfig()
}

func initConfig() Config {
	var newConfig Config
	configPath := "config/config.yaml"
	configPathEnv := os.Getenv("CONFIG_PATH")
	if configPathEnv != "" {
		configPath = configPathEnv
	}
	slog.Debug("config path", configPath)

	dir, file := filepath.Split(configPath)
	newViper := viper.New()
	newViper.SetConfigType("yaml")
	newViper.SetConfigName(strings.TrimSuffix(file, filepath.Ext(file)))
	newViper.AddConfigPath(fmt.Sprintf("./%s", dir))

	if err := newViper.ReadInConfig(); err != nil {
		panic("config read fail")
	}
	if err := newViper.Unmarshal(&newConfig); err != nil {
		panic("fail to parse config")
	}

	newConfig.setServerConfigFromEnvIfExist()

	return newConfig
}

// 環境変数からサーバーの設定値を読み込むメソッド
func (c *Config) setServerConfigFromEnvIfExist() {
	listenAddrFromEnv := os.Getenv("LISTEN_ADDR")
	if listenAddrFromEnv != "" {
		c.Server.ListenAddr = listenAddrFromEnv
		slog.Debug("listen addr from env", listenAddrFromEnv)
	}
}

func GetConfig() Config {
	return c
}
