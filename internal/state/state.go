package state

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/mxcd/gitops-cli/internal/secret"
	"github.com/mxcd/gitops-cli/internal/util"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
	"github.com/andybalholm/crlf"
)

type State struct {
	// List of secrets in the state
	Secrets []*SecretState
}

type SecretState struct {
	// unique uuid of the secret
	ID string
	// Target is the target of the secret
	Target secret.SecretTarget
	// Name of the secret
	Name string
	// Namespace of the secret
	Namespace string
	// Type of the secret
	Type string
	// Path is the path to the secret file
	Path string
	// SHA256 hash of the decrypted secret file
	BinaryDataHash string
}

var state *State

func LoadState(c *cli.Context) error {
	// Load state from project root
	stateFileName := path.Join(util.GetRootDir(), ".gitops-state.yaml")
	stats, err := os.Stat(stateFileName)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new state
			state = &State{
				Secrets: []*SecretState{},
			}
			return nil
		} else {
			return err
		}
	} 
	
	// Load state from file
	if stats.IsDir() {
		return fmt.Errorf("state file is a directory")
	}
	stateFile, err := os.Open(stateFileName)
	if err != nil {
		return err
	}
	defer func() {
		if err := stateFile.Close(); err != nil {
			panic(err)
		}
	}()
	reader := crlf.NewReader(stateFile)
	yamlFile, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &state)
	return err
}

func (s *State) Save(c *cli.Context) error {
	// Save state to project root
	stateFileName := path.Join(util.GetRootDir(), ".gitops-state.yaml")
	yamlFile, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	stateFile, err := os.Create(stateFileName)
	if err != nil {
		return err
	}
	defer func() {
		if err := stateFile.Close(); err != nil {
			panic(err)
		}
	}()
	
	writer := crlf.NewWriter(stateFile)
	_, err = writer.Write(yamlFile)
	return err
}

func (s *State) GetByPath(path string) *SecretState {
	for _, secret := range s.Secrets {
		if secret.Path == path {
			return secret
		}
	}
	return nil
}

func (s *State) Add(secret *secret.Secret) *SecretState {
	stateSecret := &SecretState{
		ID: secret.ID,
		Target: secret.Target,
		Path: secret.Path,
		BinaryDataHash: secret.BinaryDataHash,
		Name: secret.Name,
		Namespace: secret.Namespace,
		Type: secret.Type,
	}
	s.Secrets = append(s.Secrets, stateSecret)
	return stateSecret
}

// TODO prohibit update of the secret type
func (s *SecretState) Update(secret *secret.Secret) {
	secret.ID = s.ID
	s.Target = secret.Target
	s.Name = secret.Name
	s.Namespace = secret.Namespace
	s.BinaryDataHash = secret.BinaryDataHash
	s.Type = secret.Type
}

func (s *SecretState) CombinedName() string {
	return s.Namespace + "/" + s.Name
}

func (s *State) SetSecrets(secrets []*SecretState) {
	s.Secrets = secrets
}

func GetState() *State {
	return state
}