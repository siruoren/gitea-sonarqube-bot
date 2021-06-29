package sonarqube

import (
	"fmt"
	"log"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"gitea-sonarqube-pr-bot/internal/settings"
)

func inProjectsMapping(p []settings.Project, n string) bool {
	for _, proj := range p {
		if proj.SonarQube.Key == n {
			return true
		}
	}

	return false
}

func HandleWebhook(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	project := r.Header.Get("X-SonarQube-Project")
	if !inProjectsMapping(settings.Projects, project) {
		log.Printf("Received hook for project '%s' which is not configured. Request ignored.", project)

		rw.WriteHeader(http.StatusOK)
		io.WriteString(rw, fmt.Sprintf(`{"message": "Project '%s' not in configured list. Request ignored."}`, project))
		return
	}

	log.Printf("Received hook for project '%s'. Processing data.", project)

	raw, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Printf("Error reading request body %s", err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		io.WriteString(rw, fmt.Sprintf(`{"message": "%s"}`, err.Error()))
		return
	}

	w, ok := NewWebhook(raw)
	if !ok {
		rw.WriteHeader(http.StatusUnprocessableEntity)
		io.WriteString(rw, `{"message": "Error parsing POST body."}`)
		return
	}

	// Send response to SonarQube at this point to ensure being within 10 seconds limit of webhook response timeout
	rw.WriteHeader(http.StatusOK)
	io.WriteString(rw, `{"message": "Processing data. See bot logs for details."}`)

	if strings.ToLower(w.Branch.Type) != "pull_request" {
		log.Print("Ignore Hook for non-PR")
		return
	}

	log.Printf("%s", w)
}
