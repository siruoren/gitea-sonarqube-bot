package sonarqube

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"gitea-sonarqube-pr-bot/internal/settings"

	"github.com/stretchr/testify/assert"
)

type ClientMock struct {
	responseError error
	handler       http.HandlerFunc
	recoder       *httptest.ResponseRecorder
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	c.handler.ServeHTTP(c.recoder, req)

	return &http.Response{
		StatusCode: c.recoder.Code,
		Body:       c.recoder.Result().Body,
	}, c.responseError
}

func TestParsePRIndexSuccess(t *testing.T) {
	settings.Pattern = &settings.PatternConfig{
		RegExp: regexp.MustCompile(`^PR-(\d+)$`),
	}

	actual, _ := ParsePRIndex("PR-1337")
	assert.Equal(t, 1337, actual, "PR index parsing is broken")

	t.Cleanup(func() {
		settings.Pattern = nil
	})
}

func TestParsePRIndexNonIntegerFailure(t *testing.T) {
	settings.Pattern = &settings.PatternConfig{
		RegExp: regexp.MustCompile(`^PR-(\d+)$`),
	}

	_, err := ParsePRIndex("PR-invalid")
	assert.EqualErrorf(t, err, "branch name 'PR-invalid' does not match regex '^PR-(\\d+)$'", "Integer parsing succeeds unexpectedly")

	t.Cleanup(func() {
		settings.Pattern = nil
	})
}

func TestPRNameFromIndex(t *testing.T) {
	settings.Pattern = &settings.PatternConfig{
		Template: "PR-%d",
	}

	assert.Equal(t, "PR-1337", PRNameFromIndex(1337))

	t.Cleanup(func() {
		settings.Pattern = nil
	})
}

func TestGetRenderedQualityGateSuccess(t *testing.T) {
	actual := GetRenderedQualityGate("OK")

	assert.Contains(t, actual, ":white_check_mark:", "Undetected successful quality gate during status rendering")
}

func TestGetRenderedQualityGateFailure(t *testing.T) {
	actual := GetRenderedQualityGate("ERROR")

	assert.Contains(t, actual, ":x:", "Undetected failed quality gate during status rendering")
}

func TestGetPullRequestUrl(t *testing.T) {
	sdk := &SonarQubeSdk{
		baseUrl: "https://sonarqube.example.com",
	}
	settings.Pattern = &settings.PatternConfig{
		Template: "PR-%d",
	}

	actual := sdk.GetPullRequestUrl("test-project", 1337)
	assert.Equal(t, "https://sonarqube.example.com/dashboard?id=test-project&pullRequest=PR-1337", actual, "PR Dashboard URL building broken")

	t.Cleanup(func() {
		settings.Pattern = nil
	})
}

func TestRetrieveDataFromApiSuccess(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"pullRequests":[{"key":"PR-1","title":"pr-branch","branch":"pr-branch","base":"main","status":{"qualityGateStatus":"OK","bugs":0,"vulnerabilities":0,"codeSmells":0},"analysisDate":"2022-06-12T11:23:09+0000","target":"main"}]}`))
	})
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: io.ReadAll,
	}

	request := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	wrapper := &PullsResponse{}
	err := retrieveDataFromApi(sdk, request, wrapper)

	assert.Nil(t, err, "Successful data retrieval broken and throws error")
	assert.Equal(t, "Basic dGVzdC10b2tlbjo=", request.Header.Get("Authorization"), "Authorization header not set")
	assert.Equal(t, "PR-1", wrapper.PullRequests[0].Key, "Unmarshallowing into wrapper broken")
}

func TestRetrieveDataFromApiRequestError(t *testing.T) {
	expected := fmt.Errorf("This error indicates an error while performing the request")
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			recoder:       httptest.NewRecorder(),
			responseError: expected,
		},
		bodyReader: io.ReadAll,
	}

	request := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	err := retrieveDataFromApi(sdk, request, &PullsResponse{})

	assert.ErrorIs(t, err, expected, "Undetected request performing error")
}

func TestRetrieveDataFromApiUnauthorized(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder.Code = http.StatusUnauthorized
	})
	sdk := &SonarQubeSdk{
		token: "simulated-invalid-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       recorder,
			responseError: nil,
		},
		bodyReader: io.ReadAll,
	}

	request := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	err := retrieveDataFromApi(sdk, request, &PullsResponse{})

	assert.Errorf(t, err, "missing or invalid API token", "Undetected unauthorized error")
}

func TestRetrieveDataFromApiBodyReadError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"pullRequests":[{"key":"PR-1","title":"pr-branch","branch":"pr-branch","base":"main","status":{"qualityGateStatus":"OK","bugs":0,"vulnerabilities":0,"codeSmells":0},"analysisDate":"2022-06-12T11:23:09+0000","target":"main"}]}`))
	})
	expected := fmt.Errorf("Error reading body content")
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: func(r io.Reader) ([]byte, error) {
			return []byte(``), expected
		},
	}

	request := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	err := retrieveDataFromApi(sdk, request, &PullsResponse{})

	assert.ErrorIs(t, err, expected, "Undetected body processing error")
}

func TestRetrieveDataFromApiBodyUnmarshalError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"pullReq`))
	})
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: io.ReadAll,
	}

	request := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	err := retrieveDataFromApi(sdk, request, &PullsResponse{})

	assert.Errorf(t, err, "unexpected end of JSON input", "Undetected body unmarshal error")
}

func TestFetchPullRequestsSuccess(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"pullRequests":[{"key":"PR-1","title":"pr-branch","branch":"pr-branch","base":"main","status":{"qualityGateStatus":"OK","bugs":0,"vulnerabilities":0,"codeSmells":0},"analysisDate":"2022-06-12T11:23:09+0000","target":"main"}]}`))
	})
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return httptest.NewRequest(method, target, body), nil
		},
	}

	actual, err := sdk.fetchPullRequests("test-project")

	assert.Nil(t, err, "Successful data retrieval broken and throws error")
	assert.IsType(t, &PullsResponse{}, actual, "Happy path broken")
}

func TestFetchPullRequestsRequestBuildingFailure(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"pullRequests":[{"key":"PR-1","title":"pr-branch","branch":"pr-branch","base":"main","status":{"qualityGateStatus":"OK","bugs":0,"vulnerabilities":0,"codeSmells":0},"analysisDate":"2022-06-12T11:23:09+0000","target":"main"}]}`))
	})
	expected := fmt.Errorf("Some simulated error")
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return nil, expected
		},
	}

	_, err := sdk.fetchPullRequests("test-project")

	assert.Equal(t, expected, err, "Unexpected error instance returned")
}

func TestFetchPullRequestsRequestError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"pullRequests":[{"key":"PR-1","title":"pr-branch","branch":"pr-branch","base":"main","status":{"qualityGateStatus":"OK","bugs":0,"vulnerabilities":0,"codeSmells":0},"analysisDate":"2022-06-12T11:23:09+0000","target":"main"}]}`))
	})
	expected := fmt.Errorf("Some simulated error")
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: expected,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return httptest.NewRequest(method, target, body), nil
		},
	}

	_, err := sdk.fetchPullRequests("test-project")

	assert.Equal(t, expected, err)
}

func TestFetchPullRequestsErrorsInResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"errors":[{"msg":"Project 'test-project' not found"}]}`))
	})
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return httptest.NewRequest(method, target, body), nil
		},
	}

	_, err := sdk.fetchPullRequests("test-project")

	assert.Errorf(t, err, "Project 'test-project' not found", "Response error parsing broken")
}

func TestGetPullRequestSuccess(t *testing.T) {
	settings.Pattern = &settings.PatternConfig{
		Template: "PR-%d",
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"pullRequests":[{"key":"PR-1","title":"pr-branch","branch":"pr-branch","base":"main","status":{"qualityGateStatus":"OK","bugs":0,"vulnerabilities":0,"codeSmells":0},"analysisDate":"2022-06-12T11:23:09+0000","target":"main"}]}`))
	})
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return httptest.NewRequest(method, target, body), nil
		},
	}

	actual, err := sdk.GetPullRequest("test-project", 1)

	assert.Nil(t, err, "Successful data retrieval broken and throws error")
	assert.IsType(t, &PullRequest{}, actual, "Happy path broken")

	t.Cleanup(func() {
		settings.Pattern = nil
	})
}

func TestGetPullRequestFetchError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"pullRequests":[{"key":"PR-1","title":"pr-branch","branch":"pr-branch","base":"main","status":{"qualityGateStatus":"OK","bugs":0,"vulnerabilities":0,"codeSmells":0},"analysisDate":"2022-06-12T11:23:09+0000","target":"main"}]}`))
	})
	expected := fmt.Errorf("Some simulated error")
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: expected,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return httptest.NewRequest(method, target, body), nil
		},
	}

	_, err := sdk.GetPullRequest("test-project", 1)

	assert.Errorf(t, err, "fetching pull requests failed", "Incorrect edge case is throwing errors")
	assert.Errorf(t, err, "Some simulated error", "Unexpected error cause")
}

func TestGetPullRequestUnknownPR(t *testing.T) {
	settings.Pattern = &settings.PatternConfig{
		Template: "PR-%d",
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"pullRequests":[{"key":"PR-1","title":"pr-branch","branch":"pr-branch","base":"main","status":{"qualityGateStatus":"OK","bugs":0,"vulnerabilities":0,"codeSmells":0},"analysisDate":"2022-06-12T11:23:09+0000","target":"main"}]}`))
	})
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return httptest.NewRequest(method, target, body), nil
		},
	}

	_, err := sdk.GetPullRequest("test-project", 1337)

	assert.Errorf(t, err, "no pull request found with name 'PR-1337'")

	t.Cleanup(func() {
		settings.Pattern = nil
	})
}

func TestGetMeasuresSuccess(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"component":{"key":"test-project","name":"Test Project","qualifier":"TRK","measures":[{"metric":"bugs","value":"0","bestValue":true}],"pullRequest":"PR-1"},"metrics":[{"key":"bugs","name":"Bugs","description":"Bugs","domain":"Reliability","type":"INT","higherValuesAreBetter":false,"qualitative":false,"hidden":false,"custom":false,"bestValue":"0"}]}`))
	})
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return httptest.NewRequest(method, target, body), nil
		},
	}

	actual, err := sdk.GetMeasures("test-project", "PR-1")

	assert.Nil(t, err, "Successful data retrieval broken and throws error")
	assert.IsType(t, &MeasuresResponse{}, actual, "Happy path broken")
}

func TestGetMeasuresRequestBuildingFailure(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"component":{"key":"test-project","name":"Test Project","qualifier":"TRK","measures":[{"metric":"bugs","value":"0","bestValue":true}],"pullRequest":"PR-1"},"metrics":[{"key":"bugs","name":"Bugs","description":"Bugs","domain":"Reliability","type":"INT","higherValuesAreBetter":false,"qualitative":false,"hidden":false,"custom":false,"bestValue":"0"}]}`))
	})
	expected := fmt.Errorf("Some simulated error")
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return nil, expected
		},
	}

	_, err := sdk.GetMeasures("test-project", "PR-1")

	assert.Equal(t, expected, err, "Unexpected error instance returned")
}

func TestGetMeasuresRequestError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"component":{"key":"test-project","name":"Test Project","qualifier":"TRK","measures":[{"metric":"bugs","value":"0","bestValue":true}],"pullRequest":"PR-1"},"metrics":[{"key":"bugs","name":"Bugs","description":"Bugs","domain":"Reliability","type":"INT","higherValuesAreBetter":false,"qualitative":false,"hidden":false,"custom":false,"bestValue":"0"}]}`))
	})
	expected := fmt.Errorf("Some simulated error")
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: expected,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return httptest.NewRequest(method, target, body), nil
		},
	}

	_, err := sdk.GetMeasures("test-project", "PR-1")

	assert.Equal(t, expected, err)
}

func TestGetMeasuresErrorsInResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"errors":[{"msg":"Component 'non-existing-project' of pull request 'PR-1' not found"}]}`))
	})
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return httptest.NewRequest(method, target, body), nil
		},
	}

	_, err := sdk.GetMeasures("non-existing-project", "PR-1")

	assert.Errorf(t, err, "Component 'non-existing-project' of pull request 'PR-1' not found", "Response error parsing broken")
}

func TestComposeGiteaCommentSuccess(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"component":{"key":"test-project","name":"Test Project","qualifier":"TRK","measures":[{"metric":"bugs","value":"10","bestValue":false}],"pullRequest":"PR-1"},"metrics":[{"key":"bugs","name":"Bugs","description":"Bugs","domain":"Reliability","type":"INT","higherValuesAreBetter":false,"qualitative":false,"hidden":false,"custom":false,"bestValue":"0"}]}`))
	})
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return httptest.NewRequest(method, target, body), nil
		},
	}

	actual, err := sdk.ComposeGiteaComment(&CommentComposeData{
		Key:         "test-project",
		PRName:      "PR-1",
		Url:         "https://sonarqube.example.com",
		QualityGate: "OK",
	})

	assert.Nil(t, err, "Successful comment composing throwing errors")
	assert.Contains(t, actual, ":white_check_mark:", "Happy path [Quality Gate] broken")
	assert.Contains(t, actual, "| Metric | Current |", "Happy path [Metrics Header] broken")
	assert.Contains(t, actual, "| Bugs | 10 |", "Happy path [Metrics Values] broken")
	assert.Contains(t, actual, "https://sonarqube.example.com", "Happy path [Link] broken")
	assert.Contains(t, actual, "/sq-bot review", "Happy path [Command] broken")
}

func TestComposeGiteaCommentError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"component":{"key":"test-project","name":"Test Project","qualifier":"TRK","measures":[{"metric":"bugs","value":"10","bestValue":false}],"pullRequest":"PR-1"},"metrics":[{"key":"bugs","name":"Bugs","description":"Bugs","domain":"Reliability","type":"INT","higherValuesAreBetter":false,"qualitative":false,"hidden":false,"custom":false,"bestValue":"0"}]}`))
	})
	expected := fmt.Errorf("Expected error from GetMeasures")
	sdk := &SonarQubeSdk{
		token: "test-token",
		client: &ClientMock{
			handler:       handler,
			recoder:       httptest.NewRecorder(),
			responseError: nil,
		},
		bodyReader: io.ReadAll,
		httpRequest: func(method, target string, body io.Reader) (*http.Request, error) {
			return nil, expected
		},
	}

	_, err := sdk.ComposeGiteaComment(&CommentComposeData{
		Key:         "test-project",
		PRName:      "PR-1",
		Url:         "https://sonarqube.example.com",
		QualityGate: "OK",
	})

	assert.Errorf(t, err, expected.Error(), "Undetected error while composing comment")
}

func TestNew(t *testing.T) {
	settings.SonarQube = settings.SonarQubeConfig{
		Url: "http://example.com",
		Token: &settings.Token{
			Value: "test-token",
		},
	}
	assert.IsType(t, &SonarQubeSdk{}, New(), "")
}
