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
}

type SonarQubeConfig struct {
	Url string
	Token Token
	Webhook Webhook
}

type Project struct {
	SonarQube struct {
		Key string
	} `mapstructure:"sonarqube"`
	Gitea GiteaRepository
}

var (
	Gitea GiteaConfig
	SonarQube SonarQubeConfig
	Projects []Project
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

	viper.SetDefault("sonarqube.url", "")
	viper.SetDefault("sonarqube.token.value", "")
	viper.SetDefault("sonarqube.token.file", "")
	viper.SetDefault("sonarqube.webhook.secret", "")
	viper.SetDefault("sonarqube.webhook.secretFile", "")

	viper.SetDefault("projects", []Project{})
}

func ReadSecretFile(file string, defaultValue string) (string) {
	if file == "" {
		return defaultValue
	}

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
		Projects []Project
	}

	err = viper.Unmarshal(&fullConfig)
	if err != nil {
		panic(fmt.Errorf("Unable to load config into struct, %v", err))
	}

	if len(fullConfig.Projects) == 0 {
		panic("Invalid configuration. At least one project mapping is necessary.")
	}

	Gitea = fullConfig.Gitea
	SonarQube = fullConfig.SonarQube
	Projects = fullConfig.Projects

	Gitea.Webhook.Secret = ReadSecretFile(Gitea.Webhook.SecretFile, Gitea.Webhook.Secret)
	Gitea.Token.Value = ReadSecretFile(Gitea.Token.File, Gitea.Token.Value)
	SonarQube.Webhook.Secret = ReadSecretFile(SonarQube.Webhook.SecretFile, SonarQube.Webhook.Secret)
	SonarQube.Token.Value = ReadSecretFile(SonarQube.Token.File, SonarQube.Token.Value)
}
