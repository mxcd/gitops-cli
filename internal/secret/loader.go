package secret

import (
	"github.com/mxcd/gitops-cli/internal/util"
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