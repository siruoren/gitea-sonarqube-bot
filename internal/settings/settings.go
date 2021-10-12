package settings

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

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
	v.SetDefault("sonarqube.additionalMetrics", []string{})
	v.SetDefault("projects", []Project{})

	return v
}

func Load(configPath string) {
	r := newConfigReader()
	r.AddConfigPath(configPath)

	err := r.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error while reading config file: %w", err))
	}

	var projects []Project

	err = r.UnmarshalKey("projects", &projects)
	if err != nil {
		panic(fmt.Errorf("unable to load project mapping: %s", err.Error()))
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
		Url:               r.GetString("sonarqube.url"),
		Token:             NewToken(r, "sonarqube", errCallback),
		Webhook:           NewWebhook(r, "sonarqube", errCallback),
		AdditionalMetrics: r.GetStringSlice("sonarqube.additionalMetrics"),
	}
}
