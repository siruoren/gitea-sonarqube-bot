package settings

type GiteaRepository struct {
	Owner string
	Name  string
}

type giteaConfig struct {
	Url     string
	Token   *Token
	Webhook *Webhook
}
