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

func NewConfigReader() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config.yaml")
	v.SetConfigType("yaml")
	v.SetEnvPrefix("prbot")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	v.SetDefault("gitea.url", "")
	v.SetDefault("gitea.token.value", "")
	v.SetDefault("gitea.token.file", "")
	v.SetDefault("gitea.webhook.secret", "")
	v.SetDefault("gitea.webhook.secretFile", "")
	v.SetDefault("sonarqube.url", "")
	v.SetDefault("sonarqube.token.value", "")
	v.SetDefault("sonarqube.token.file", "")
	v.SetDefault("sonarqube.webhook.secret", "")
	v.SetDefault("sonarqube.webhook.secretFile", "")
	v.SetDefault("projects", []Project{})

	return v
}

func Load(configPath string) {
	r := NewConfigReader()
	r.AddConfigPath(configPath)

	err := r.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error while reading config file: %w \n", err))
	}

	var fullConfig struct {
		Gitea GiteaConfig
		SonarQube SonarQubeConfig `mapstructure:"sonarqube"`
		Projects []Project
	}

	err = r.Unmarshal(&fullConfig)
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
