package state

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/mxcd/gitops-cli/internal/secret"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
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
	// Path is the path to the secret file
	Path string
	// SHA256 hash of the decrypted secret file
	BinaryDataHash string
}

var state *State

func LoadState(c *cli.Context) error {
	// Load state from project root
	stateFileName := path.Join(c.String("root-dir"), ".gitops-state.yaml")
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
	yamlFile, err := ioutil.ReadFile(stateFileName)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &state)
	return err
}

func (s *State) Save(c *cli.Context) error {
	// Save state to project root
	stateFileName := path.Join(c.String("root-dir"), ".gitops-state.yaml")
	yamlFile, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(stateFileName, yamlFile, 0644)
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
	}
	s.Secrets = append(s.Secrets, stateSecret)
	return stateSecret
}

func (s *SecretState) Update(secret *secret.Secret) {
	secret.ID = s.ID
	s.Target = secret.Target
	s.Name = secret.Name
	s.Namespace = secret.Namespace
	s.BinaryDataHash = secret.BinaryDataHash
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