package settings

type GiteaRepository struct {
	Owner string
	Name  string
}

type GiteaConfig struct {
	Url     string
	Token   *Token
	Webhook *Webhook
}
