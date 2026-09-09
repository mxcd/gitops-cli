package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mxcd/gitops-cli/internal/patch"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// fakePatcher records the calls it receives and returns preconfigured results.
type fakePatcher struct {
	PatchCalls      [][]patch.PatchTask
	PatchBatchCalls []patch.PatchBatch
	CommitHash      string
	Err             error
}

func (f *fakePatcher) Prepare(options *patch.PrepareOptions) error {
	return nil
}

func (f *fakePatcher) Patch(patchTasks []patch.PatchTask) error {
	f.PatchCalls = append(f.PatchCalls, patchTasks)
	return f.Err
}

func (f *fakePatcher) PatchBatch(batch patch.PatchBatch) (string, error) {
	f.PatchBatchCalls = append(f.PatchBatchCalls, batch)
	if f.Err != nil {
		return "", f.Err
	}
	return f.CommitHash, nil
}

func newTestServer(t *testing.T, patcher patch.PatchMethod) *Server {
	server, err := NewServer(&RouterOptions{
		DevMode:    true,
		Port:       0,
		ApiBaseUrl: "/api/v1",
		ApiKeys:    []string{"test-api-key"},
	}, patcher)
	assert.NoError(t, err)
	assert.NotNil(t, server)

	server.RegisterMiddlewares()
	err = server.RegisterRoutes()
	assert.NoError(t, err)

	return server
}

// executeRequest sends a request to the server and returns the recorder.
func executeRequest(t *testing.T, server *Server, method string, url string, body string, apiKey string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	recorder := httptest.NewRecorder()
	server.Engine.ServeHTTP(recorder, req)

	return recorder
}

const validBatchBody = `{
  "actor": "ci-bot",
  "files": [
    {"filePath": "applications/dev/service-a/values.yaml", "patches": [{"selector": ".service.image.tag", "value": "v1.0.0"}]},
    {"filePath": "applications/dev/service-b/values.yaml", "patches": [{"selector": ".service.image.tag", "value": "v1.0.0"}]}
  ]
}`

func TestPatchesHandlerSuccess(t *testing.T) {
	patcher := &fakePatcher{CommitHash: "abc123"}
	server := newTestServer(t, patcher)

	recorder := executeRequest(t, server, http.MethodPut, "/api/v1/patches", validBatchBody, "test-api-key")
	assert.Equal(t, 200, recorder.Code)

	var response map[string]any
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["message"])
	assert.Equal(t, "abc123", response["commit"])

	assert.Len(t, patcher.PatchBatchCalls, 1)
	batch := patcher.PatchBatchCalls[0]
	assert.Equal(t, "ci-bot", batch.Actor)
	assert.Len(t, batch.Files, 2)
	assert.Equal(t, "applications/dev/service-a/values.yaml", batch.Files[0].FilePath)
	assert.Equal(t, "applications/dev/service-b/values.yaml", batch.Files[1].FilePath)
	assert.Equal(t, ".service.image.tag", batch.Files[1].Patches[0].Selector)
}

func TestPatchesHandlerNoChanges(t *testing.T) {
	patcher := &fakePatcher{CommitHash: ""}
	server := newTestServer(t, patcher)

	recorder := executeRequest(t, server, http.MethodPut, "/api/v1/patches", validBatchBody, "test-api-key")
	assert.Equal(t, 200, recorder.Code)

	var response map[string]any
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["message"])
	assert.Equal(t, "", response["commit"])
}

func TestPatchesHandlerInvalidRequests(t *testing.T) {
	testCases := []struct {
		name          string
		body          string
		expectedError string
	}{
		{name: "invalid json", body: `{"files": [`, expectedError: "invalid input"},
		{name: "empty files", body: `{"files": []}`, expectedError: "files"},
		{name: "missing files", body: `{"actor": "ci-bot"}`, expectedError: "files"},
		{name: "traversing file path", body: `{"files": [{"filePath": "../escape.yaml", "patches": [{"selector": ".a", "value": "b"}]}]}`, expectedError: "escape the repository root"},
		{name: "absolute file path", body: `{"files": [{"filePath": "/etc/passwd", "patches": [{"selector": ".a", "value": "b"}]}]}`, expectedError: "relative to the repository root"},
		{name: "file without patches", body: `{"files": [{"filePath": "a/values.yaml", "patches": []}]}`, expectedError: "patches"},
		{name: "patch without selector", body: `{"files": [{"filePath": "a/values.yaml", "patches": [{"selector": "", "value": "b"}]}]}`, expectedError: "selector"},
		{
			name:          "duplicate file paths",
			body:          `{"files": [{"filePath": "a/values.yaml", "patches": [{"selector": ".a", "value": "b"}]}, {"filePath": "a/./values.yaml", "patches": [{"selector": ".a", "value": "b"}]}]}`,
			expectedError: "duplicate",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			patcher := &fakePatcher{CommitHash: "abc123"}
			server := newTestServer(t, patcher)

			recorder := executeRequest(t, server, http.MethodPut, "/api/v1/patches", testCase.body, "test-api-key")
			assert.Equal(t, 400, recorder.Code)
			assert.Contains(t, recorder.Body.String(), testCase.expectedError)
			assert.Empty(t, patcher.PatchBatchCalls, "the patcher must not be called for invalid input")
		})
	}
}

func TestPatchesHandlerPatcherError(t *testing.T) {
	patcher := &fakePatcher{Err: errors.New("git exploded")}
	server := newTestServer(t, patcher)

	recorder := executeRequest(t, server, http.MethodPut, "/api/v1/patches", validBatchBody, "test-api-key")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "error executing patching")
	// the internal error must not leak to the client
	assert.NotContains(t, recorder.Body.String(), "git exploded")
	assert.Len(t, patcher.PatchBatchCalls, 1)
}

func TestPatchesHandlerPatcherValidationError(t *testing.T) {
	patcher := &fakePatcher{Err: patch.ErrInvalidPatchBatch}
	server := newTestServer(t, patcher)

	recorder := executeRequest(t, server, http.MethodPut, "/api/v1/patches", validBatchBody, "test-api-key")
	assert.Equal(t, 400, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "invalid patch batch")
}

func TestPatchesHandlerAuthentication(t *testing.T) {
	testCases := []struct {
		name   string
		apiKey string
	}{
		{name: "no api key", apiKey: ""},
		{name: "wrong api key", apiKey: "wrong-api-key"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			patcher := &fakePatcher{CommitHash: "abc123"}
			server := newTestServer(t, patcher)

			recorder := executeRequest(t, server, http.MethodPut, "/api/v1/patches", validBatchBody, testCase.apiKey)
			assert.Equal(t, 401, recorder.Code)
			assert.Empty(t, patcher.PatchBatchCalls)
		})
	}
}

func TestPatchRouteStillWorks(t *testing.T) {
	patcher := &fakePatcher{}
	server := newTestServer(t, patcher)

	body := `{"actor": "ci-bot", "filePath": "applications/dev/service-a/values.yaml", "patches": [{"selector": ".service.image.tag", "value": "v1.0.0"}]}`
	recorder := executeRequest(t, server, http.MethodPut, "/api/v1/patch", body, "test-api-key")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"message":"ok"`)

	assert.Len(t, patcher.PatchCalls, 1)
	assert.Len(t, patcher.PatchCalls[0], 1)
	assert.Equal(t, "applications/dev/service-a/values.yaml", patcher.PatchCalls[0][0].FilePath)
	assert.Empty(t, patcher.PatchBatchCalls)
}

func TestHealthRouteIsUnprotected(t *testing.T) {
	server := newTestServer(t, &fakePatcher{})

	recorder := executeRequest(t, server, http.MethodGet, "/api/v1/health", "", "")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"status":"ok"`)
}
