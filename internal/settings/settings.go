package settings

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/viper"
)

type GiteaRepository struct {
	Owner string
	Name string
}

type WebhookSecret struct {
	Value string
	File string
}

type GiteaConfig struct {
	Url string
	Token string
	WebhookSecret WebhookSecret `mapstructure:"webhookSecret"`
	Repositories []GiteaRepository
}

type SonarQubeConfig struct {
	Url string
	Token string
	WebhookSecret WebhookSecret `mapstructure:"webhookSecret"`
	Projects []string
}

var (
	Gitea GiteaConfig
	SonarQube SonarQubeConfig
)

func init() {
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("prbot")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()

	ApplyConfigDefaults()
}

func ApplyConfigDefaults() {
	viper.SetDefault("gitea.url", "")
	viper.SetDefault("gitea.token", "")
	viper.SetDefault("gitea.webhookSecret.value", "")
	viper.SetDefault("gitea.webhookSecret.file", "")
	viper.SetDefault("gitea.repositories", []interface{}{})
	viper.SetDefault("sonarqube.url", "")
	viper.SetDefault("sonarqube.token", "")
	viper.SetDefault("sonarqube.webhookSecret.value", "")
	viper.SetDefault("sonarqube.webhookSecret.file", "")
	viper.SetDefault("sonarqube.projects", []string{})
}

func ReadSecretFile(file string) string {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		panic(fmt.Errorf("Cannot read '%s' or it is no regular file. %w", file, err))
	}

	return string(content)
}

func Load(configPath string) {
	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error while reading config file: %w \n", err))
	}

	var fullConfig struct {
		Gitea GiteaConfig
		SonarQube SonarQubeConfig `mapstructure:"sonarqube"`
	}

	err = viper.Unmarshal(&fullConfig)
	if err != nil {
		panic(fmt.Errorf("Unable to load config into struct, %v", err))
	}

	Gitea = fullConfig.Gitea
	SonarQube = fullConfig.SonarQube

	if Gitea.WebhookSecret.File != "" {
		Gitea.WebhookSecret.Value = ReadSecretFile(Gitea.WebhookSecret.File)
	}

	if SonarQube.WebhookSecret.File != "" {
		SonarQube.WebhookSecret.Value = ReadSecretFile(SonarQube.WebhookSecret.File)
	}
}
