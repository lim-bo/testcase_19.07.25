package settings

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

var (
	once     sync.Once
	instance *Config
)

type Config struct {
}

func GetConfig() *Config {
	once.Do(func() {
		viper.AddConfigPath("../../config/cfg.yaml")
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatal(err)
		}
		instance = &Config{}
	})
	return instance
}

func (cfg *Config) GetString(key string) string {
	return viper.GetString(key)
}

func (cfg *Config) Get(key string) any {
	return viper.Get(key)
}
