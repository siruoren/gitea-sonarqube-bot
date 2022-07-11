package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	giteaSdk "gitea-sonarqube-bot/internal/clients/gitea"
	sqSdk "gitea-sonarqube-bot/internal/clients/sonarqube"
	"gitea-sonarqube-bot/internal/settings"
	webhook "gitea-sonarqube-bot/internal/webhooks/sonarqube"
)

type SonarQubeWebhookHandlerInferface interface {
	Handle(r *http.Request) (int, string)
}

type SonarQubeWebhookHandler struct {
	giteaSdk giteaSdk.GiteaSdkInterface
	sqSdk    sqSdk.SonarQubeSdkInterface
}

func (*SonarQubeWebhookHandler) inProjectsMapping(p []settings.Project, n string) (bool, int) {
	for idx, proj := range p {
		if proj.SonarQube.Key == n {
			return true, idx
		}
	}

	return false, 0
}

func (h *SonarQubeWebhookHandler) processData(w *webhook.Webhook, repo settings.GiteaRepository) {
	status := giteaSdk.StatusOK
	if w.QualityGate.Status != "OK" {
		status = giteaSdk.StatusFailure
	}
	_ = h.giteaSdk.UpdateStatus(repo, w.GetRevision(), giteaSdk.StatusDetails{
		Url:     w.Branch.Url,
		Message: w.QualityGate.Status,
		State:   status,
	})

	comment, err := h.sqSdk.ComposeGiteaComment(&sqSdk.CommentComposeData{
		Key:         w.Project.Key,
		PRName:      w.Branch.Name,
		Url:         w.Branch.Url,
		QualityGate: w.QualityGate.Status,
	})
	if err != nil {
		return
	}
	h.giteaSdk.PostComment(repo, w.PRIndex, comment)
}

func (h *SonarQubeWebhookHandler) Handle(r *http.Request) (int, string) {
	projectName := r.Header.Get("X-SonarQube-Project")
	found, pIdx := h.inProjectsMapping(settings.Projects, projectName)
	if !found {
		log.Printf("Received hook for project '%s' which is not configured. Request ignored.", projectName)
		return http.StatusOK, fmt.Sprintf("Project '%s' not in configured list. Request ignored.", projectName)
	}

	log.Printf("Received hook for project '%s'. Processing data.", projectName)

	if r.Body != nil {
		defer r.Body.Close()
	}

	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body %s", err.Error())
		return http.StatusInternalServerError, err.Error()
	}

	ok, err := isValidWebhook(raw, settings.SonarQube.Webhook.Secret, r.Header.Get("X-Sonar-Webhook-HMAC-SHA256"), "SonarQube")
	if !ok {
		log.Print(err.Error())
		return http.StatusPreconditionFailed, "Webhook validation failed. Request rejected."
	}

	w, ok := webhook.New(raw)
	if !ok {
		return http.StatusUnprocessableEntity, "Error parsing POST body."
	}

	if strings.ToLower(w.Branch.Type) != "pull_request" {
		log.Println("Ignore Hook for non-PR analysis")
		return http.StatusOK, "Ignore Hook for non-PR analysis."
	}

	h.processData(w, settings.Projects[pIdx].Gitea)

	return http.StatusOK, "Processing data. See bot logs for details."
}

func NewSonarQubeWebhookHandler(g giteaSdk.GiteaSdkInterface, sq sqSdk.SonarQubeSdkInterface) SonarQubeWebhookHandlerInferface {
	return &SonarQubeWebhookHandler{
		giteaSdk: g,
		sqSdk:    sq,
	}
}
