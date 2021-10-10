package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	giteaSdk "gitea-sonarqube-pr-bot/internal/clients/gitea"
	sqSdk "gitea-sonarqube-pr-bot/internal/clients/sonarqube"
	webhook "gitea-sonarqube-pr-bot/internal/webhooks/gitea"
)

type GiteaWebhookHandler struct {
	giteaSdk giteaSdk.GiteaSdkInterface
	sqSdk    sqSdk.SonarQubeSdkInterface
}

func (h *GiteaWebhookHandler) parseBody(rw http.ResponseWriter, r *http.Request) ([]byte, error) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	raw, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("Error reading request body %s", err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		io.WriteString(rw, fmt.Sprintf(`{"message": "%s"}`, err.Error()))
		return nil, err
	}

	return raw, nil
}

func (h *GiteaWebhookHandler) HandleSynchronize(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	_, err := h.parseBody(rw, r)
	if err != nil {
		return
	}

	rw.WriteHeader(http.StatusOK)
	io.WriteString(rw, `{"message": "Processing data. See bot logs for details."}`)
}

func (h *GiteaWebhookHandler) HandleComment(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	raw, err := h.parseBody(rw, r)
	if err != nil {
		return
	}

	w, ok := webhook.New(raw)
	if !ok {
		rw.WriteHeader(http.StatusUnprocessableEntity)
		io.WriteString(rw, `{"message": "Error parsing POST body."}`)
		return
	}

	if err := w.Validate(); err != nil {
		rw.WriteHeader(http.StatusOK)
		io.WriteString(rw, fmt.Sprintf(`{"message": "%s"}`, err.Error()))
		return
	}

	rw.WriteHeader(http.StatusOK)
	io.WriteString(rw, `{"message": "Processing data. See bot logs for details."}`)

	w.ProcessData(h.giteaSdk, h.sqSdk)
}

func NewGiteaWebhookHandler(g giteaSdk.GiteaSdkInterface, sq sqSdk.SonarQubeSdkInterface) *GiteaWebhookHandler {
	return &GiteaWebhookHandler{
		giteaSdk: g,
		sqSdk:    sq,
	}
}
