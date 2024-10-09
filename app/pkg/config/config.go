package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppConfig struct {
		HttpAddr           string `env:"HTTP_ADDR" envDefault:":8085"`
		GinMode            string `env:"GIN_MODE"`
		LiteserverPineIP   int64  `env:"LITESERVER_PINE_IP"`
		LiteserverPinePort int    `env:"LITESERVER_PINE_PORT"`
		LiteserverPineType string `env:"LITESERVER_PINE_TYPE"`
		LiteserverPineKey  string `env:"LITESERVER_PINE_KEY"`
	}
	Redis struct {
		Host     string `env:"REDIS_HOST"`
		Port     string `env:"REDIS_PORT"`
		Password string `env:"REDIS_PASSWORD"`
		Size     int    `env:"REDIS_SIZE"`
	}
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		log.Print("gather config")

		instance = &Config{}

		if err := cleanenv.ReadConfig(".envs/.env", instance); err != nil {
			helpText := "Rpine Demo Aggregation"
			help, _ := cleanenv.GetDescription(instance, &helpText)
			log.Print(help)
			log.Fatal(err)
		}
	})
	return instance
}
