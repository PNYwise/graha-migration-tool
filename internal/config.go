package internal

import (
	"os"

	"github.com/spf13/viper"
)

func New() *viper.Viper {
	confName := os.Getenv("APP_ENV")
	if confName == "" {
		confName = "development"
	}

	conf := viper.New()
	conf.SetConfigName(confName)
	conf.SetConfigType("yaml")
	conf.AddConfigPath("./conf/")
	err := conf.ReadInConfig()
	if err != nil {
		panic(err)
	}
	return conf
}