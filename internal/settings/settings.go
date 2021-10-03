package settings

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type GiteaRepository struct {
	Owner string
	Name  string
}

type giteaConfig struct {
	Url     string
	Token   *token
	Webhook *webhook
}

type sonarQubeConfig struct {
	Url     string
	Token   *token
	Webhook *webhook
}

type Project struct {
	SonarQube struct {
		Key string
	} `mapstructure:"sonarqube"`
	Gitea GiteaRepository
}

var (
	Gitea     giteaConfig
	SonarQube sonarQubeConfig
	Projects  []Project
)

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

	var projects []Project

	err = r.UnmarshalKey("projects", &projects)
	if err != nil {
		panic(fmt.Errorf("Unable to load project mapping: %s", err.Error()))
	}

	if len(projects) == 0 {
		panic("Invalid configuration. At least one project mapping is necessary.")
	}

	Projects = projects

	errCallback := func(msg string) { panic(msg) }

	Gitea = giteaConfig{
		Url:     r.GetString("gitea.url"),
		Token:   NewToken(r, "gitea", errCallback),
		Webhook: NewWebhook(r, "gitea", errCallback),
	}
	SonarQube = sonarQubeConfig{
		Url:     r.GetString("sonarqube.url"),
		Token:   NewToken(r, "sonarqube", errCallback),
		Webhook: NewWebhook(r, "sonarqube", errCallback),
	}
}
