package secret

import (
	"strings"

	"github.com/mxcd/gitops-cli/internal/util"
	log "github.com/sirupsen/logrus"
)

/*
	Loads all the secrets from the local file system
	Applies the specified target filter
	Use SecretTargetAll to load all secrets
*/
func LoadLocalSecrets(targetFilter SecretTarget) ([]*Secret, error) {
	secretFileNames, err := util.GetSecretFiles()
	if err != nil {
		return nil, err
	}
	secrets := []*Secret{}
	for _, secretFileName := range secretFileNames {
		if strings.HasSuffix(secretFileName, "values.gitops.secret.enc.yml") || strings.HasSuffix(secretFileName, "values.gitops.secret.enc.yaml")  {
			log.Trace("Skipping values file: ", secretFileName)
			continue
		}
		secret, err := FromPath(secretFileName)
		if err != nil {
			return nil, err
		}
		if secret.Target == targetFilter || targetFilter == SecretTargetAll {
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}