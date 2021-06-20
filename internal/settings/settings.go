package settings

import (
	"fmt"

	"github.com/spf13/viper"
)

type GiteaRepository struct {
	Owner string
	Name string
}

type GiteaConfig struct {
	Url string
	Token string
	WebhookSecret string `mapstructure:"webhookSecret"`
	Repositories []GiteaRepository
}

var (
	Gitea GiteaConfig
)

func Load(configPath string) {
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error while reading config file: %w \n", err))
	}

	var giteaConfig GiteaConfig

	if viper.Sub("gitea") == nil {
		panic("Gitea not configured")
	}

	err = viper.UnmarshalKey("gitea", &giteaConfig)
	if err != nil {
		panic(fmt.Errorf("Unable to decode into struct, %v", err))
	}

	Gitea = giteaConfig
}
