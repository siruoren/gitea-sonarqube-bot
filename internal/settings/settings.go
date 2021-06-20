package settings

import (
	"fmt"
	"strings"

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

func init() {
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("prbot")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()
}

func Load(configPath string) {
	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error while reading config file: %w \n", err))
	}

	if viper.IsSet("gitea") == false {
		panic("Gitea not configured")
	}

	var fullConfig struct {
		Gitea GiteaConfig
	}

	err = viper.Unmarshal(&fullConfig)
	if err != nil {
		panic(fmt.Errorf("Unable to load config into struct, %v", err))
	}

	Gitea = fullConfig.Gitea
}
