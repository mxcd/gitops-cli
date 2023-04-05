package secret

import (
	"bytes"
	"encoding/hex"
	"errors"
	"path"
	"text/template"

	"crypto/sha256"

	"github.com/mxcd/gitops-cli/internal/templating"
	"github.com/mxcd/gitops-cli/internal/util"
	log "github.com/sirupsen/logrus"
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
	BinaryDataHash string

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

type TemplateData struct {
	Values map[interface{}]interface{}
}

func (s *Secret) Load() error {
	if s.Path == "" {
		return errors.New("secret path is empty")
	}

	absoluteSecretPath := path.Join(util.GetRootDir(), s.Path)
	decryptedFileContent, err := util.DecryptFile(absoluteSecretPath)
	if err != nil {
		return err
	}

	// execute templating on the secret file data
	data := TemplateData{
		Values: templating.GetValuesForPath(s.Path),
	}
	stringData := string(decryptedFileContent)
	tmpl, err := template.New(s.Path).Parse(stringData)
	if err != nil {
		log.Error("Error parsing template for secret " + s.Path)
		return err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, data)
	if err != nil {
		log.Error("Error executing template for secret " + s.Path)
		return err
	}

	s.BinaryData = buf.Bytes()

	binaryHash := sha256.Sum256(s.BinaryData)
	hash := binaryHash[:]
	s.BinaryDataHash = hex.EncodeToString(hash)

	var secretFile SecretFile
	yaml.UnmarshalStrict(s.BinaryData, &secretFile)

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

	if util.GetCliContext().Bool("print") {
		s.PrettyPrint()
	}
	
	return nil
}

func (s *Secret) CombinedName() string {
	return s.Namespace + "/" + s.Name
}

func (s *Secret) PrettyPrint() {
	cleartext := util.GetCliContext().Bool("cleartext")
	println("---")
	println(s.CombinedName())
	println("  target: " + string(s.Target))
	println("  type: " + s.Type)
	println("  data:")
	for k, v := range s.Data {
		if cleartext {
			println("    " + k + ":")
			println(v)
		} else {
			println("    " + k + ": " + "********")
		}
	}
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

