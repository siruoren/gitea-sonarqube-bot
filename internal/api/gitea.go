package api

import (
	"io/ioutil"
	"log"
	"net/http"

	giteaSdk "codeberg.org/justusbunsi/gitea-sonarqube-bot/internal/clients/gitea"
	sqSdk "codeberg.org/justusbunsi/gitea-sonarqube-bot/internal/clients/sonarqube"
	"codeberg.org/justusbunsi/gitea-sonarqube-bot/internal/settings"
	webhook "codeberg.org/justusbunsi/gitea-sonarqube-bot/internal/webhooks/gitea"
)

type GiteaWebhookHandlerInferface interface {
	HandleSynchronize(r *http.Request) (int, string)
	HandleComment(r *http.Request) (int, string)
}

type GiteaWebhookHandler struct {
	giteaSdk giteaSdk.GiteaSdkInterface
	sqSdk    sqSdk.SonarQubeSdkInterface
}

func (h *GiteaWebhookHandler) parseBody(r *http.Request) ([]byte, error) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	raw, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("Error reading request body %s", err.Error())
		return nil, err
	}

	return raw, nil
}

func (h *GiteaWebhookHandler) HandleSynchronize(r *http.Request) (int, string) {
	raw, err := h.parseBody(r)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}

	ok, err := isValidWebhook(raw, settings.Gitea.Webhook.Secret, r.Header.Get("X-Gitea-Signature"), "Gitea")
	if !ok {
		log.Print(err.Error())
		return http.StatusPreconditionFailed, "Webhook validation failed. Request rejected."
	}

	w, ok := webhook.NewPullWebhook(raw)
	if !ok {
		return http.StatusUnprocessableEntity, "Error parsing POST body."
	}

	if err := w.Validate(); err != nil {
		return http.StatusOK, err.Error()
	}

	w.ProcessData(h.giteaSdk, h.sqSdk)

	return http.StatusOK, "Processing data. See bot logs for details."
}

func (h *GiteaWebhookHandler) HandleComment(r *http.Request) (int, string) {
	raw, err := h.parseBody(r)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}

	ok, err := isValidWebhook(raw, settings.Gitea.Webhook.Secret, r.Header.Get("X-Gitea-Signature"), "Gitea")
	if !ok {
		log.Print(err.Error())
		return http.StatusPreconditionFailed, "Webhook validation failed. Request rejected."
	}

	w, ok := webhook.NewCommentWebhook(raw)
	if !ok {
		return http.StatusUnprocessableEntity, "Error parsing POST body."
	}

	if err := w.Validate(); err != nil {
		return http.StatusOK, err.Error()
	}

	w.ProcessData(h.giteaSdk, h.sqSdk)

	return http.StatusOK, "Processing data. See bot logs for details."
}

func NewGiteaWebhookHandler(g giteaSdk.GiteaSdkInterface, sq sqSdk.SonarQubeSdkInterface) GiteaWebhookHandlerInferface {
	return &GiteaWebhookHandler{
		giteaSdk: g,
		sqSdk:    sq,
	}
}
