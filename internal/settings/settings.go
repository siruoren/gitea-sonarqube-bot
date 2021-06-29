package settings

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/viper"
)

type giteaRepository struct {
	Owner string
	Name string
}

type token struct {
	Value string
	File string
}

type webhook struct {
	Secret string
	SecretFile string
}

type giteaConfig struct {
	Url string
	Token token
	Webhook webhook
}

type sonarQubeConfig struct {
	Url string
	Token token
	Webhook webhook
}

type Project struct {
	SonarQube struct {
		Key string
	} `mapstructure:"sonarqube"`
	Gitea giteaRepository
}

type fullConfig struct {
	Gitea giteaConfig
	SonarQube sonarQubeConfig `mapstructure:"sonarqube"`
	Projects []Project
}

var (
	Gitea giteaConfig
	SonarQube sonarQubeConfig
	Projects []Project
)

func readSecretFile(file string, defaultValue string) (string) {
	if file == "" {
		return defaultValue
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		panic(fmt.Errorf("Cannot read '%s' or it is no regular file. %w", file, err))
	}

	return string(content)
}

func newConfigReader() *viper.Viper {
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
	r := newConfigReader()
	r.AddConfigPath(configPath)

	err := r.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error while reading config file: %w \n", err))
	}

	var configuration fullConfig

	err = r.Unmarshal(&configuration)
	if err != nil {
		panic(fmt.Errorf("Unable to load config into struct, %v", err))
	}

	if len(configuration.Projects) == 0 {
		panic("Invalid configuration. At least one project mapping is necessary.")
	}

	Gitea = configuration.Gitea
	SonarQube = configuration.SonarQube
	Projects = configuration.Projects

	Gitea.Webhook.Secret = readSecretFile(Gitea.Webhook.SecretFile, Gitea.Webhook.Secret)
	Gitea.Token.Value = readSecretFile(Gitea.Token.File, Gitea.Token.Value)
	SonarQube.Webhook.Secret = readSecretFile(SonarQube.Webhook.SecretFile, SonarQube.Webhook.Secret)
	SonarQube.Token.Value = readSecretFile(SonarQube.Token.File, SonarQube.Token.Value)
}
