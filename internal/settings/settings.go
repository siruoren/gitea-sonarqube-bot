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

type Token struct {
	Value string
	File string
}

type Webhook struct {
	Secret string
	SecretFile string
}

type GiteaConfig struct {
	Url string
	Token Token
	Webhook Webhook
	Repositories []GiteaRepository
}

type SonarQubeConfig struct {
	Url string
	Token Token
	Webhook Webhook
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
	viper.SetDefault("gitea.token.value", "")
	viper.SetDefault("gitea.token.file", "")
	viper.SetDefault("gitea.webhook.secret", "")
	viper.SetDefault("gitea.webhook.secretFile", "")
	viper.SetDefault("gitea.repositories", []GiteaRepository{})

	viper.SetDefault("sonarqube.url", "")
	viper.SetDefault("sonarqube.token.value", "")
	viper.SetDefault("sonarqube.token.file", "")
	viper.SetDefault("sonarqube.webhook.secret", "")
	viper.SetDefault("sonarqube.webhook.secretFile", "")
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

	if Gitea.Webhook.SecretFile != "" {
		Gitea.Webhook.Secret = ReadSecretFile(Gitea.Webhook.SecretFile)
	}

	if SonarQube.Webhook.SecretFile != "" {
		SonarQube.Webhook.Secret = ReadSecretFile(SonarQube.Webhook.SecretFile)
	}
}
