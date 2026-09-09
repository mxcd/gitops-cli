package patch

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type recordedRequest struct {
	Method      string
	Path        string
	ApiKey      string
	ContentType string
	Body        []byte
}

// newRecordingServer starts a test server that records the incoming request
// and answers with the given status code and body.
func newRecordingServer(t *testing.T, statusCode int, responseBody string) (*httptest.Server, *recordedRequest) {
	recorded := &recordedRequest{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		recorded.Method = r.Method
		recorded.Path = r.URL.Path
		recorded.ApiKey = r.Header.Get("X-API-Key")
		recorded.ContentType = r.Header.Get("Content-Type")
		recorded.Body = body

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, err = w.Write([]byte(responseBody))
		assert.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	return server, recorded
}

func TestRepositoryServerPatcherPatchBatch(t *testing.T) {
	server, recorded := newRecordingServer(t, http.StatusOK, `{"message":"ok","commit":"abc123"}`)

	patcher := &RepositoryServerPatcher{
		RepositoryServerURL:    server.URL,
		RepositoryServerApiKey: "test-api-key",
	}

	batch := PatchBatch{
		Actor: "ci-bot",
		Files: []FilePatch{
			{FilePath: "applications/dev/service-a/values.yaml", Patches: []Patch{{Selector: ".service.image.tag", Value: "v1.0.0"}}},
			{FilePath: "applications/dev/service-b/values.yaml", Patches: []Patch{{Selector: ".service.image.tag", Value: "v1.0.0"}}},
		},
	}

	commit, err := patcher.PatchBatch(batch)
	assert.NoError(t, err)
	assert.Equal(t, "abc123", commit)

	assert.Equal(t, http.MethodPut, recorded.Method)
	assert.Equal(t, "/patches", recorded.Path)
	assert.Equal(t, "test-api-key", recorded.ApiKey)
	assert.Equal(t, "application/json", recorded.ContentType)

	var sentBatch PatchBatch
	err = json.Unmarshal(recorded.Body, &sentBatch)
	assert.NoError(t, err)
	assert.Equal(t, "ci-bot", sentBatch.Actor)
	assert.Len(t, sentBatch.Files, 2, "all files of the batch must be sent")
	assert.Equal(t, "applications/dev/service-a/values.yaml", sentBatch.Files[0].FilePath)
	assert.Equal(t, "applications/dev/service-b/values.yaml", sentBatch.Files[1].FilePath)
	assert.Equal(t, ".service.image.tag", sentBatch.Files[1].Patches[0].Selector)
}

func TestRepositoryServerPatcherPatchBatchNoFiles(t *testing.T) {
	server, recorded := newRecordingServer(t, http.StatusOK, `{"message":"ok","commit":"abc123"}`)

	patcher := &RepositoryServerPatcher{
		RepositoryServerURL:    server.URL,
		RepositoryServerApiKey: "test-api-key",
	}

	commit, err := patcher.PatchBatch(PatchBatch{})
	assert.NoError(t, err)
	assert.Empty(t, commit)
	assert.Empty(t, recorded.Method, "no request must be sent for an empty batch")
}

func TestRepositoryServerPatcherPatchBatchServerError(t *testing.T) {
	server, _ := newRecordingServer(t, http.StatusInternalServerError, `{"error":"error executing patching"}`)

	patcher := &RepositoryServerPatcher{
		RepositoryServerURL:    server.URL,
		RepositoryServerApiKey: "test-api-key",
	}

	commit, err := patcher.PatchBatch(PatchBatch{
		Files: []FilePatch{{FilePath: "applications/dev/service-a/values.yaml", Patches: []Patch{{Selector: ".service.image.tag", Value: "v1.0.0"}}}},
	})
	assert.Error(t, err)
	assert.Empty(t, commit)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "error executing patching")
}

func TestRepositoryServerPatcherPatchUsesSingleFileEndpoint(t *testing.T) {
	server, recorded := newRecordingServer(t, http.StatusOK, `{"message":"ok"}`)

	patcher := &RepositoryServerPatcher{
		RepositoryServerURL:    server.URL,
		RepositoryServerApiKey: "test-api-key",
	}

	err := patcher.Patch([]PatchTask{{
		Actor:    "ci-bot",
		FilePath: "applications/dev/service-a/values.yaml",
		Patches:  []Patch{{Selector: ".service.image.tag", Value: "v1.0.0"}},
	}})
	assert.NoError(t, err)

	assert.Equal(t, http.MethodPut, recorded.Method)
	assert.Equal(t, "/patch", recorded.Path)

	var sentTask PatchTask
	err = json.Unmarshal(recorded.Body, &sentTask)
	assert.NoError(t, err)
	assert.Equal(t, "applications/dev/service-a/values.yaml", sentTask.FilePath)
}
