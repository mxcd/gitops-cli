package secret

import (
	"errors"

	"github.com/mxcd/gitops-cli/internal/util"
	"gopkg.in/yaml.v2"
)

type Secret struct {
	// Path is the path to the secret file
	Path string

	// Target is the target of the secret
	Target SecretFileTarget

	// Type is the type of the secret
	Type string

	// calculated name of the secret
	Name string

	// optional namespace of the secret
	Namespace string

	// Data is the decrypted data from the secret file
	Data map[string]string
}

type SecretFileTarget string
var SecretFileTargetVault SecretFileTarget = "vault"
var SecretFileTargetKubernetes SecretFileTarget = "k8s"

type SecretFile struct {
	Target     SecretFileTarget  `yaml:"target"`
	Name       string            `yaml:"name,omitempty"`
	Namespace  string            `yaml:"namespace" default:"default"`
	Type       string            `yaml:"type" default:"Opaque"`
	Data			 map[string]string `yaml:"data"`
}

func (s *Secret) Load() error {
	if s.Path == "" {
		return errors.New("secret path is empty")
	}

	decryptedFileContent, err := util.DecryptFile(s.Path)
	if err != nil {
		return err
	}

	var secretFile SecretFile
	yaml.UnmarshalStrict(decryptedFileContent, &secretFile)

	s.Target = secretFile.Target
	
	if secretFile.Name != "" {
		s.Name = secretFile.Name
	} else {
		// basename of path
		s.Name = util.GetSecretBasename(s.Path)
	}

	if secretFile.Namespace != "" {
		s.Namespace = secretFile.Namespace
	} else {
		s.Namespace = "default"
	}

	if secretFile.Type != "" {
		s.Type = secretFile.Type
	} else {
		s.Type = "Opaque"
	}
		
	s.Data = secretFile.Data
	
	return nil
}