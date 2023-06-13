package secret

import (
	"errors"
	"strings"

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
	for _, secretFileName := range secretFileNames {
		bar.Add(1)
		secret, err := FromPath(secretFileName)
		if err != nil {
			bar.Finish()
			return nil, err
		}
		if secret.TargetType != targetTypeFilter && targetTypeFilter != SecretTargetTypeAll {
			log.Trace("Skipping file due to targetType filter: ", secretFileName)
			continue
		}
		if clusterLimit != "" && secret.Target != clusterLimit {
			log.Trace("Skipping file due to target filter: ", secretFileName)
			continue
		}
		for _, s := range secrets {
			if s.Name == secret.Name && s.Target == secret.Target {
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
