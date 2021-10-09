package settings

type Project struct {
	SonarQube struct {
		Key string
	} `mapstructure:"sonarqube"`
	Gitea GiteaRepository
}
