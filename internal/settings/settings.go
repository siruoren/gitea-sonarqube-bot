package settings

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

var (
	Gitea     GiteaConfig
	SonarQube SonarQubeConfig
	Projects  []Project
	Pattern   *PatternConfig
)

func newConfigReader(configFile string) *viper.Viper {
	v := viper.New()
	v.SetConfigFile(configFile)
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
	v.SetDefault("namingPattern.regex", `^PR-(\d+)$`)
	v.SetDefault("namingPattern.template", "PR-%d")

	return v
}

func Load(configFile string) {
	r := newConfigReader(configFile)

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

	Gitea = GiteaConfig{
		Url:     r.GetString("gitea.url"),
		Token:   NewToken(r.GetString, "gitea", errCallback),
		Webhook: NewWebhook(r.GetString, "gitea", errCallback),
	}
	SonarQube = SonarQubeConfig{
		Url:               r.GetString("sonarqube.url"),
		Token:             NewToken(r.GetString, "sonarqube", errCallback),
		Webhook:           NewWebhook(r.GetString, "sonarqube", errCallback),
		AdditionalMetrics: r.GetStringSlice("sonarqube.additionalMetrics"),
	}
	Pattern = &PatternConfig{
		RegExp:   regexp.MustCompile(r.GetString("namingPattern.regex")),
		Template: r.GetString("namingPattern.template"),
	}
}
