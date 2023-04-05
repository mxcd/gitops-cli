package secret

import (
	"fmt"

	color "github.com/TwiN/go-color"
	"github.com/mxcd/gitops-cli/internal/util"
)

type SecretDiffType string
var SecretDiffTypeUnchanged SecretDiffType = "unchanged"
var SecretDiffTypeAdded SecretDiffType = "added"
var SecretDiffTypeRemoved SecretDiffType = "removed"
var SecretDiffTypeChanged SecretDiffType = "changed"

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

func CompareSecrets(oldSecret *Secret, newSecret *Secret) *SecretDiff {
	diffEntries := []SecretDiffEntry{}
	
	// if both secrets are nil, they are equal
	if oldSecret == nil && newSecret == nil {
		return &SecretDiff{
			Equal: true,
			Type: SecretDiffTypeUnchanged,
			Name: "",
			Namespace: "",
			Entries: diffEntries,
		}
	}

	// if the new secret is not nil and the old secret is nil, the new secret is added
	if oldSecret == nil && newSecret != nil {
		for key, value := range newSecret.Data {
			diffEntries = append(diffEntries, SecretDiffEntry{
				Type: SecretDiffTypeAdded,
				Key: fmt.Sprintf("data.%s", key),
				OldValue: "",
				NewValue: value,
				Sensitive: true,
			})
		}
		return &SecretDiff{
			Equal: false,
			Type: SecretDiffTypeAdded,
			Name: newSecret.Name,
			Namespace: newSecret.Namespace,
			Entries: diffEntries,
		}
	}

	// if the old secret is not nil and the new secret is nil, the old secret is removed
	if oldSecret != nil && newSecret == nil {
		for key, value := range oldSecret.Data {
			diffEntries = append(diffEntries, SecretDiffEntry{
				Type: SecretDiffTypeRemoved,
				Key: fmt.Sprintf("data.%s", key),
				OldValue: value,
				NewValue: "",
				Sensitive: true,
			})
		}
		return &SecretDiff{
			Equal: false,
			Type: SecretDiffTypeRemoved,
			Name: oldSecret.Name,
			Namespace: oldSecret.Namespace,
			Entries: diffEntries,
		}
	}
	

	if oldSecret.Target != newSecret.Target {
		diffEntries = append(diffEntries, SecretDiffEntry{
			Type: SecretDiffTypeChanged,
			Key: "target",
			OldValue: string(oldSecret.Target),
			NewValue: string(newSecret.Target),
			Sensitive: false,
		})
	}

	if oldSecret.Name != newSecret.Name {
		diffEntries = append(diffEntries, SecretDiffEntry{
			Type: SecretDiffTypeChanged,
			Key: "name",
			OldValue: oldSecret.Name,
			NewValue: newSecret.Name,
			Sensitive: false,
		})
	}

	if oldSecret.Namespace != newSecret.Namespace {
		diffEntries = append(diffEntries, SecretDiffEntry{
			Type: SecretDiffTypeChanged,
			Key: "namespace",
			OldValue: oldSecret.Namespace,
			NewValue: newSecret.Namespace,
			Sensitive: false,
		})
	}

	if oldSecret.Type != newSecret.Type {
		diffEntries = append(diffEntries, SecretDiffEntry{
			Type: SecretDiffTypeChanged,
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
				Type: SecretDiffTypeRemoved,
				Key: fmt.Sprintf("data.%s", key),
				OldValue: value,
				NewValue: "",
				Sensitive: true,
			})
		} else {
			// check if value is equal
			if value != newSecret.Data[key] {
				diffEntries = append(diffEntries, SecretDiffEntry{
					Type: SecretDiffTypeChanged,
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
				Type: SecretDiffTypeAdded,
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
		diff.Type = SecretDiffTypeChanged;
		diff.Equal = false
	} else {
		diff.Type = SecretDiffTypeUnchanged;
		diff.Equal = true
	}

	return &diff
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
			if entry.Sensitive && !util.GetCliContext().Bool("cleartext") {
				safeOldValue = util.ToRedactedString(entry.OldValue)
				safeNewValue = util.ToRedactedString(entry.NewValue)
			}
			switch entry.Type {
			case SecretDiffTypeAdded:
				println(color.Ize(color.Green, fmt.Sprintf("  + %s: %s", entry.Key, safeNewValue)))
			case SecretDiffTypeRemoved:
				println(color.Ize(color.Red, fmt.Sprintf("  - %s: %s", entry.Key, safeOldValue)))
			case SecretDiffTypeChanged:
				println(color.Ize(color.Yellow, fmt.Sprintf("  ~ %s: %s => %s", entry.Key, safeOldValue, safeNewValue)))
			}
		}
	}

	if d.Equal {
		println(color.InGray(combinedSecretName), color.InGray(": "), color.InBold(color.InGray("unchanged")))
		return
	}
	switch d.Type {
		case SecretDiffTypeAdded:
			println(color.InGreen(combinedSecretName), color.InGreen(": "), color.InBold(color.InGreen("add")))
		case SecretDiffTypeRemoved:
			println(color.InRed(combinedSecretName), color.InRed(": "), color.InBold(color.InRed("remove")))
		case SecretDiffTypeChanged:
			println(color.InYellow(combinedSecretName), color.InYellow(": "), color.InBold(color.InYellow("change")))
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