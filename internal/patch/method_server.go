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

func (p *RepositoryServerPatcher) Patch(patchTasks []PatchTask) error {
	if len(patchTasks) == 0 {
		log.Warn().Msg("No patch tasks provided, skipping patching")
		return nil
	}

	if len(patchTasks) > 1 {
		log.Warn().Msg("More than one patch task provided, only the first one will be applied")
	}

	jsonData, err := json.Marshal(patchTasks[0])
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal patch tasks")
		return fmt.Errorf("failed to marshal patch tasks: %w", err)
	}
	log.Debug().Msgf("Patch task JSON: %s", string(jsonData))

	requestURL := fmt.Sprintf("%s/patch", p.RepositoryServerURL)
	log.Debug().Msgf("Request URL: %s", requestURL)

	req, err := http.NewRequest(http.MethodPut, requestURL, bytes.NewReader(jsonData))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create HTTP request")
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.RepositoryServerApiKey)

	log.Info().Msg("Sending patch request to repository server")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send request to repository server")
		return fmt.Errorf("failed to send request to repository server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read response body")
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Error().Msgf("Repository server returned status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("repository server returned status %d: %s", resp.StatusCode, string(body))
	}

	log.Info().Msg("Patch applied successfully via repository server.")
	return nil
}
