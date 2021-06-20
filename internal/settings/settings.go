package settings

import (
	"fmt"

	"github.com/spf13/viper"
)

func Load(configPath string) {
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()

  if err != nil {
  	panic(fmt.Errorf("Fatal error while reading config file: %w \n", err))
  }
}
