package sonarqube

type PullRequest struct {
	Key    string `json:"key"`
	Status struct {
		QualityGateStatus string `json:"qualityGateStatus"`
	} `json:"status"`
}

type PullsResponse struct {
	PullRequests []PullRequest `json:"pullRequests"`
	Errors       []Error       `json:"errors"`
}

func (r *PullsResponse) GetPullRequest(name string) *PullRequest {
	for _, pr := range r.PullRequests {
		if pr.Key == name {
			return &pr
		}
	}
	return nil
}
