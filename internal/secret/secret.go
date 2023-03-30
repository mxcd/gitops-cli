package secret

import (
	"errors"

	"crypto/sha256"

	"github.com/mxcd/gitops-cli/internal/util"
	"gopkg.in/yaml.v2"
)

type Secret struct {
	// unique uuid of the secret
	ID string
	// Path is the path to the secret file
	Path string

	// Target is the target of the secret
	Target SecretTarget

	// Type is the type of the secret
	Type string

	// calculated name of the secret
	Name string

	// optional namespace of the secret
	Namespace string

	// Decrypted binary data from the secret file
	BinaryData []byte

	// SHA256 hash of the decrypted binary data
	BinaryDataHash [32]byte

	// Data is the decrypted data from the secret file
	Data map[string]string
}

type SecretTarget string
var SecretTargetVault SecretTarget = "vault"
var SecretTargetKubernetes SecretTarget = "k8s"
var SecretTargetAll SecretTarget = "all"

type SecretFile struct {
	Target     SecretTarget      `yaml:"target"`
	Name       string            `yaml:"name,omitempty"`
	Namespace  string            `yaml:"namespace" default:"default"`
	Type       string            `yaml:"type" default:"Opaque"`
	Data			 map[string]string `yaml:"data"`
	ID         string            `yaml:"id,omitempty"`
}

func (s *Secret) Load() error {
	if s.Path == "" {
		return errors.New("secret path is empty")
	}

	decryptedFileContent, err := util.DecryptFile(s.Path)
	if err != nil {
		return err
	}

	s.BinaryData = decryptedFileContent
	s.BinaryDataHash = sha256.Sum256(decryptedFileContent)

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

func FromPath(path string) (*Secret, error) {
	s := Secret {
		Path: path,
	}

	err := s.Load()
	if err != nil {
		return nil, err
	}

	return &s, nil
}

