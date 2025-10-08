package secret

import (
	"errors"
	"strings"
	"sync"

	"github.com/mxcd/gitops-cli/internal/util"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

/*
Loads all the secrets from the local file system
Applies the specified target filter
Use SecretTargetAll to load all secrets
*/
func LoadLocalSecrets(targetTypeFilter SecretTargetType) ([]*Secret, error) {
	return LoadLocalSecretsLimited(targetTypeFilter, "", "")
}

func LoadLocalSecretsLimited(targetTypeFilter SecretTargetType, directoryLimit string, clusterLimit string) ([]*Secret, error) {
	// retrieve all secret files
	secretFileNames, err := util.GetSecretFiles()
	if err != nil {
		return nil, err
	}

	// Filter by directory limit
	filteredFileNames := []string{}
	for _, secretFileName := range secretFileNames {
		if !strings.HasPrefix(secretFileName, directoryLimit) {
			log.Trace("Skipping file due to directory filter: ", secretFileName)
			continue
		}
		if strings.HasSuffix(secretFileName, "values.gitops.secret.enc.yml") || strings.HasSuffix(secretFileName, "values.gitops.secret.enc.yaml") {
			log.Trace("Skipping values file: ", secretFileName)
			continue
		}
		filteredFileNames = append(filteredFileNames, secretFileName)
	}
	secretFileNames = filteredFileNames

	parallelism := util.GetCliContext().Int("parallelism")
	if parallelism < 1 {
		parallelism = 1
	}

	secrets := []*Secret{}
	bar := progressbar.NewOptions(len(secretFileNames),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetElapsedTime(false),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetDescription("[green][Loading local secrets][reset]"),
	)

	// Create channels for parallel processing
	type secretResult struct {
		secret *Secret
		err    error
	}
	
	secretChan := make(chan string, len(secretFileNames))
	resultChan := make(chan secretResult, len(secretFileNames))
	
	// Start worker goroutines
	var workerGroup sync.WaitGroup
	for i := 0; i < parallelism; i++ {
		workerGroup.Add(1)
		go func() {
			defer workerGroup.Done()
			for secretFileName := range secretChan {
				secret, err := FromPath(secretFileName)
				resultChan <- secretResult{secret: secret, err: err}
			}
		}()
	}
	
	// Send work to workers
	go func() {
		for _, secretFileName := range secretFileNames {
			secretChan <- secretFileName
		}
		close(secretChan)
	}()
	
	// Wait for all workers to finish and close result channel
	go func() {
		workerGroup.Wait()
		close(resultChan)
	}()
	
	// Collect results
	for result := range resultChan {
		bar.Add(1)
		if result.err != nil {
			bar.Finish()
			return nil, result.err
		}
		
		secret := result.secret
		if secret.TargetType != targetTypeFilter && targetTypeFilter != SecretTargetTypeAll {
			log.Trace("Skipping file due to targetType filter: ", secret.Path)
			continue
		}
		if clusterLimit != "" && secret.Target != clusterLimit {
			log.Trace("Skipping file due to target filter: ", secret.Path)
			continue
		}
		for _, s := range secrets {
			if s.Name == secret.Name && s.Target == secret.Target && s.Namespace == secret.Namespace {
				bar.Finish()
				println("")
				log.Error("Unable to load secret '", secret.Name, "' from '", secret.Path, "' because a secret with the same name and target already exists: '", s.Path, "'")
				return nil, errors.New("error loading secrets: duplicate secret name and target")
			}
		}
		secrets = append(secrets, secret)
	}
	bar.Finish()
	println("")
	println("")
	return secrets, nil
}
