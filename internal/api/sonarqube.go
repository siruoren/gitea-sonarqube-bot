package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	giteaSdk "gitea-sonarqube-pr-bot/internal/clients/gitea"
	sqSdk "gitea-sonarqube-pr-bot/internal/clients/sonarqube"
	"gitea-sonarqube-pr-bot/internal/settings"
	webhook "gitea-sonarqube-pr-bot/internal/webhooks/sonarqube"
)

type SonarQubeWebhookHandlerInferface interface {
	Handle(rw http.ResponseWriter, r *http.Request)
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

func (h *SonarQubeWebhookHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	projectName := r.Header.Get("X-SonarQube-Project")
	found, pIdx := h.inProjectsMapping(settings.Projects, projectName)
	if !found {
		log.Printf("Received hook for project '%s' which is not configured. Request ignored.", projectName)

		rw.WriteHeader(http.StatusOK)
		io.WriteString(rw, fmt.Sprintf(`{"message": "Project '%s' not in configured list. Request ignored."}`, projectName))
		return
	}

	log.Printf("Received hook for project '%s'. Processing data.", projectName)

	if r.Body != nil {
		defer r.Body.Close()
	}

	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body %s", err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		io.WriteString(rw, fmt.Sprintf(`{"message": "%s"}`, err.Error()))
		return
	}

	ok, err := isValidWebhook(raw, settings.SonarQube.Webhook.Secret, r.Header.Get("X-Sonar-Webhook-HMAC-SHA256"), "SonarQube")
	if !ok {
		log.Print(err.Error())
		rw.WriteHeader(http.StatusPreconditionFailed)
		io.WriteString(rw, `{"message": "Webhook validation failed. Request rejected."}`)
		return
	}

	w, ok := webhook.New(raw)
	if !ok {
		rw.WriteHeader(http.StatusUnprocessableEntity)
		io.WriteString(rw, `{"message": "Error parsing POST body."}`)
		return
	}

	// Send response to SonarQube at this point to ensure being within 10 seconds limit of webhook response timeout
	rw.WriteHeader(http.StatusOK)

	if strings.ToLower(w.Branch.Type) != "pull_request" {
		io.WriteString(rw, `{"message": "Ignore Hook for non-PR analysis."}`)
		log.Println("Ignore Hook for non-PR analysis")
		return
	}

	io.WriteString(rw, `{"message": "Processing data. See bot logs for details."}`)

	h.processData(w, settings.Projects[pIdx].Gitea)
}

func NewSonarQubeWebhookHandler(g giteaSdk.GiteaSdkInterface, sq sqSdk.SonarQubeSdkInterface) SonarQubeWebhookHandlerInferface {
	return &SonarQubeWebhookHandler{
		giteaSdk: g,
		sqSdk:    sq,
	}
}
