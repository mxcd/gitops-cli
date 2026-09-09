package patch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type RepositoryServerPatcher struct {
	RepositoryServerURL    string
	RepositoryServerApiKey string
}

func NewRepoServerPatcher(c *cli.Context) (*RepositoryServerPatcher, error) {
	repositoryServerURL := c.String("repository-server")
	repositoryServerApiKey := c.String("repository-server-api-key")
	if repositoryServerURL == "" {
		log.Error().Msg("Repository server URL not provided")
		return nil, errors.New("repository server URL not provided")
	}
	if repositoryServerApiKey == "" {
		log.Error().Msg("Repository server API key not provided")
		return nil, errors.New("repository server API key not provided")
	}

	log.Debug().Msgf("Using repository server URL: %s", repositoryServerURL)

	return &RepositoryServerPatcher{
		RepositoryServerURL:    repositoryServerURL,
		RepositoryServerApiKey: repositoryServerApiKey,
	}, nil
}

func (p *RepositoryServerPatcher) Prepare(options *PrepareOptions) error {
	log.Debug().Msg("RepoServerPatcher: Prepare method called, no action required")
	return nil
}

// patchesResponse is the response body of the repository server patch
// endpoints.
type patchesResponse struct {
	Message string `json:"message"`
	Commit  string `json:"commit"`
}

// doPut marshals the payload and sends it to the given repository server path.
func (p *RepositoryServerPatcher) doPut(path string, payload any) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal patch payload")
		return nil, fmt.Errorf("failed to marshal patch payload: %w", err)
	}
	log.Debug().Msgf("Patch payload JSON: %s", string(jsonData))

	requestURL := fmt.Sprintf("%s%s", p.RepositoryServerURL, path)
	log.Debug().Msgf("Request URL: %s", requestURL)

	req, err := http.NewRequest(http.MethodPut, requestURL, bytes.NewReader(jsonData))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.RepositoryServerApiKey)

	log.Info().Msg("Sending patch request to repository server")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send request to repository server")
		return nil, fmt.Errorf("failed to send request to repository server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read response body")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Error().Msgf("Repository server returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("repository server returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (p *RepositoryServerPatcher) Patch(patchTasks []PatchTask) error {
	if len(patchTasks) == 0 {
		log.Warn().Msg("No patch tasks provided, skipping patching")
		return nil
	}

	if len(patchTasks) > 1 {
		log.Warn().Msg("More than one patch task provided, only the first one will be applied")
	}

	if _, err := p.doPut("/patch", patchTasks[0]); err != nil {
		return err
	}

	log.Info().Msg("Patch applied successfully via repository server.")
	return nil
}

// PatchBatch sends all files of the batch to the repository server so they are
// applied in a single commit. It returns the commit id reported by the server.
func (p *RepositoryServerPatcher) PatchBatch(batch PatchBatch) (string, error) {
	// validate locally so an obviously invalid batch never reaches the server
	if err := ValidatePatchBatch(batch); err != nil {
		return "", err
	}

	body, err := p.doPut("/patches", batch)
	if err != nil {
		return "", err
	}

	var response patchesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal response body")
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	log.Info().Msgf("Batch of %d file(s) applied successfully via repository server. Commit: %s", len(batch.Files), response.Commit)
	return response.Commit, nil
}
