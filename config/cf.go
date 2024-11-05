package config

import (
	"log"

	"github.com/fastdevops/monitorx/utils"

	"github.com/spf13/viper"
)

func InitConfigSet() *utils.Config {
	// init config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	var config utils.Config
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Not read config file: %v", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Not analysis config file: %v", err)
	}

	return &config
}
