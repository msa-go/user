package config

import (
	// golang package
	"log"

	// external package
	"github.com/spf13/viper"
)

func LoadConfig() Config {
	var cfg Config

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./files/config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("error read config file: %v", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("error unmarshal config: %v", err)
	}

	return cfg

}
