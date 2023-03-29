package secret

import (
	"errors"
	"fmt"

	"crypto/sha256"

	color "github.com/TwiN/go-color"
	"github.com/mxcd/gitops-cli/internal/util"
	"gopkg.in/yaml.v2"
)

type Secret struct {
	// unique uuid of the secret
	ID string
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

	// Decrypted binary data from the secret file
	BinaryData []byte

	// SHA256 hash of the decrypted binary data
	BinaryDataHash [32]byte

	// Data is the decrypted data from the secret file
	Data map[string]string
}

type SecretDiffType string
var SecretDiffEntryUnchanged SecretDiffType = "unchanged"
var SecretDiffEntryAdded SecretDiffType = "added"
var SecretDiffEntryRemoved SecretDiffType = "removed"
var SecretDiffEntryChanged SecretDiffType = "changed"

type SecretDiffEntry struct {
	Type SecretDiffType
	Key string
	OldValue string
	NewValue string
	Sensitive bool
}
type SecretDiff struct {
	// True if two compared secrets are equal
	Equal bool

	// Type of the overall secret change
	Type SecretDiffType

	// Name of the concerned secret
	Name string
	// Namespace of the concerned secret
	Namespace string

	// List of differences between two compared secrets
	Entries []SecretDiffEntry
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

func CompareSecrets(oldSecret *Secret, newSecret *Secret) SecretDiff {
	diffEntries := []SecretDiffEntry{}
	
	// if both secrets are nil, they are equal
	if oldSecret == nil && newSecret == nil {
		return SecretDiff{
			Equal: true,
			Type: SecretDiffEntryUnchanged,
			Name: "",
			Namespace: "",
			Entries: diffEntries,
		}
	}

	// if the new secret is not nil and the old secret is nil, the new secret is added
	if oldSecret == nil && newSecret != nil {
		for key, value := range newSecret.Data {
			diffEntries = append(diffEntries, SecretDiffEntry{
				Type: SecretDiffEntryAdded,
				Key: fmt.Sprintf("data.%s", key),
				OldValue: "",
				NewValue: value,
				Sensitive: true,
			})
		}
		return SecretDiff{
			Equal: false,
			Type: SecretDiffEntryAdded,
			Name: newSecret.Name,
			Namespace: newSecret.Namespace,
			Entries: diffEntries,
		}
	}

	// if the old secret is not nil and the new secret is nil, the old secret is removed
	if oldSecret != nil && newSecret == nil {
		for key, value := range oldSecret.Data {
			diffEntries = append(diffEntries, SecretDiffEntry{
				Type: SecretDiffEntryRemoved,
				Key: fmt.Sprintf("data.%s", key),
				OldValue: value,
				NewValue: "",
				Sensitive: true,
			})
		}
		return SecretDiff{
			Equal: false,
			Type: SecretDiffEntryRemoved,
			Name: oldSecret.Name,
			Namespace: oldSecret.Namespace,
			Entries: diffEntries,
		}
	}
	

	if oldSecret.Target != newSecret.Target {
		diffEntries = append(diffEntries, SecretDiffEntry{
			Type: SecretDiffEntryChanged,
			Key: "target",
			OldValue: string(oldSecret.Target),
			NewValue: string(newSecret.Target),
			Sensitive: false,
		})
	}

	if oldSecret.Name != newSecret.Name {
		diffEntries = append(diffEntries, SecretDiffEntry{
			Type: SecretDiffEntryChanged,
			Key: "name",
			OldValue: oldSecret.Name,
			NewValue: newSecret.Name,
			Sensitive: false,
		})
	}

	if oldSecret.Namespace != newSecret.Namespace {
		diffEntries = append(diffEntries, SecretDiffEntry{
			Type: SecretDiffEntryChanged,
			Key: "namespace",
			OldValue: oldSecret.Namespace,
			NewValue: newSecret.Namespace,
			Sensitive: false,
		})
	}

	if oldSecret.Type != newSecret.Type {
		diffEntries = append(diffEntries, SecretDiffEntry{
			Type: SecretDiffEntryChanged,
			Key: "type",
			OldValue: oldSecret.Type,
			NewValue: newSecret.Type,
			Sensitive: false,
		})
	}

	for key, value := range oldSecret.Data {
		// check if key is in new secret
		if _, ok := newSecret.Data[key]; !ok {
			diffEntries = append(diffEntries, SecretDiffEntry{
				Type: SecretDiffEntryRemoved,
				Key: fmt.Sprintf("data.%s", key),
				OldValue: value,
				NewValue: "",
				Sensitive: true,
			})
		} else {
			// check if value is equal
			if value != newSecret.Data[key] {
				diffEntries = append(diffEntries, SecretDiffEntry{
					Type: SecretDiffEntryChanged,
					Key: fmt.Sprintf("data.%s", key),
					OldValue: value,
					NewValue: newSecret.Data[key],
					Sensitive: true,
				})
			}
		}
	}

	for key, value := range newSecret.Data {
		// check if key is in old secret
		if _, ok := oldSecret.Data[key]; !ok {
			diffEntries = append(diffEntries, SecretDiffEntry{
				Type: SecretDiffEntryAdded,
				Key: fmt.Sprintf("data.%s", key),
				OldValue: "",
				NewValue: value,
				Sensitive: true,
			})
		}
	}

	var diffName = ""
	if oldSecret != nil {
		diffName = oldSecret.Name
	} else if newSecret != nil{
		diffName = newSecret.Name
	}

	var diffNamespace = ""
	if oldSecret != nil {
		diffNamespace = oldSecret.Namespace
	} else if newSecret != nil{
		diffNamespace = newSecret.Namespace
	}

	diff := SecretDiff {
		Name: diffName,
		Namespace: diffNamespace,
		Entries: diffEntries,
	}
	
	if len(diffEntries) > 0 {
		diff.Type = SecretDiffEntryChanged;
		diff.Equal = false
	} else {
		diff.Type = SecretDiffEntryUnchanged;
		diff.Equal = true
	}

	return diff
}

func (d *SecretDiff) Print() {
	combinedSecretName := d.Name
	if d.Namespace != "" {
		combinedSecretName = fmt.Sprintf("%s/%s", d.Namespace, d.Name)
	}

	printDetailedChanges := func() {
		for _, entry := range d.Entries {
			safeOldValue := entry.OldValue
			safeNewValue := entry.NewValue
			if entry.Sensitive {
				safeOldValue = util.ToRedactedString(entry.OldValue)
				safeNewValue = util.ToRedactedString(entry.NewValue)
			}
			switch entry.Type {
			case SecretDiffEntryAdded:
				println(color.Ize(color.Green, fmt.Sprintf("  + %s: %s", entry.Key, safeNewValue)))
			case SecretDiffEntryRemoved:
				println(color.Ize(color.Red, fmt.Sprintf("  - %s: %s", entry.Key, safeOldValue)))
			case SecretDiffEntryChanged:
				println(color.Ize(color.Yellow, fmt.Sprintf("  ~ %s: %s => %s", entry.Key, safeOldValue, safeNewValue)))
			}
		}
	}

	if d.Equal {
		println(color.Ize(color.Green, fmt.Sprintf("%s: no changes", combinedSecretName)))
		return
	}
	switch d.Type {
		case SecretDiffEntryAdded:
			println(color.Ize(color.Green, fmt.Sprintf("%s: added", combinedSecretName)))
		case SecretDiffEntryRemoved:
			println(color.Ize(color.Red, fmt.Sprintf("%s: removed", combinedSecretName)))
		case SecretDiffEntryChanged:
			println(color.Ize(color.Yellow, fmt.Sprintf("%s: changed", combinedSecretName)))
	}
	printDetailedChanges()
}

func (d *SecretDiff) GetEntry(key string) *SecretDiffEntry {
	for _, entry := range d.Entries {
		if entry.Key == key {
			return &entry
		}
	}
	return nil
}