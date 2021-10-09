package settings

type sonarQubeConfig struct {
	Url     string
	Token   *token
	Webhook *webhook
}
