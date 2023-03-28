package secret

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadSecret1(t *testing.T) {
	f := filepath.Join("..", "..", "test_assets", "test.gitops.secret.enc.yml")
	secret := Secret {
		Path: f,
	}
	err := secret.Load()

	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, secret.Target, SecretFileTargetVault, "Target should be vault")
	assert.Equal(t, secret.Name, "my-explicitly-named-secret", "Name should be my-explicitly-named-secret")
	assert.Equal(t, secret.Namespace, "default", "Namespace should be default")
	assert.Equal(t, secret.Type, "Opaque", "Type should be Opaque")

	t.Log(secret)
}

func TestLoadSecret2(t *testing.T) {
	f := filepath.Join("..", "..", "test_assets", "implicit-name.gitops.secret.enc.yml")
	secret := Secret {
		Path: f,
	}
	err := secret.Load()

	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, SecretFileTargetKubernetes, secret.Target, "Target should be k8s")
	assert.Equal(t, "implicit-name", secret.Name, "Name should be implicit-name")
	assert.Equal(t, "default", secret.Namespace, "Namespace should be default")
	assert.Equal(t, "kubernetes.io/dockerconfigjson", secret.Type, "Type should be dockerconfigjson")
}

func TestSecretComparisonTarget(t *testing.T) {
	a := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {},
	}

	b := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetVault,
		Data: map[string]string {},
	}

	diff := CompareSecrets(a, b)
	diff.Print()
	assert.Equal(t, false, diff.Equal, "Secrets should not be equal")
	assert.Equal(t, SecretDiffEntryChanged, diff.Type, "Diff type should be changed")
	assert.Equal(t, 1, len(diff.Entries), "Diff should have 1 entry")
	assert.Equal(t, SecretDiffEntryChanged, diff.Entries[0].Type, "DiffEntry type should be changed")
	assert.Equal(t, "target", diff.Entries[0].Key, "DiffEntry key should be type")
}

func TestSecretComparisonType(t *testing.T) {
	a := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {},
	}

	b := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: ".dockerconfigjson",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {},
	}

	diff := CompareSecrets(a, b)
	diff.Print()
	assert.Equal(t, false, diff.Equal, "Secrets should not be equal")
	assert.Equal(t, SecretDiffEntryChanged, diff.Type, "Diff type should be changed")
	assert.Equal(t, 1, len(diff.Entries), "Diff should have 1 entry")
	assert.Equal(t, SecretDiffEntryChanged, diff.Entries[0].Type, "DiffEntry type should be changed")
	assert.Equal(t, "type", diff.Entries[0].Key, "DiffEntry key should be type")
}

func TestSecretComparisonChangedName(t *testing.T) {
	a := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {},
	}

	b := &Secret {
		Name: "myNameExtended",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {},
	}

	diff := CompareSecrets(a, b)
	diff.Print()
	assert.Equal(t, false, diff.Equal, "Secrets should not be equal")
	assert.Equal(t, SecretDiffEntryChanged, diff.Type, "Diff type should be changed")
	assert.Equal(t, 1, len(diff.Entries), "Diff should have 1 entry")
	assert.Equal(t, SecretDiffEntryChanged, diff.Entries[0].Type, "DiffEntry type should be changed")
	assert.Equal(t, "name", diff.Entries[0].Key, "DiffEntry key should be type")
}

func TestSecretComparisonChangedNamespace(t *testing.T) {
	a := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {},
	}

	b := &Secret {
		Name: "myName",
		Namespace: "myNamespaceExtended",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {},
	}

	diff := CompareSecrets(a, b)
	diff.Print()
	assert.Equal(t, false, diff.Equal, "Secrets should not be equal")
	assert.Equal(t, SecretDiffEntryChanged, diff.Type, "Diff type should be changed")
	assert.Equal(t, 1, len(diff.Entries), "Diff should have 1 entry")
	assert.Equal(t, SecretDiffEntryChanged, diff.Entries[0].Type, "DiffEntry type should be changed")
	assert.Equal(t, "namespace", diff.Entries[0].Key, "DiffEntry key should be type")
}

func TestSecretComparisonAddData(t *testing.T) {
	a := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {},
	}

	b := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {
			"key1": "value1",
			"key2": "value2",
		},
	}

	diff := CompareSecrets(a, b)
	diff.Print()
	assert.Equal(t, false, diff.Equal, "Secrets should not be equal")
	assert.Equal(t, SecretDiffEntryChanged, diff.Type, "Diff type should be changed")
	assert.Equal(t, 2, len(diff.Entries), "Diff should have 2 entry")
	assert.Equal(t, SecretDiffEntryAdded, diff.Entries[0].Type, "DiffEntry type should be changed")
	assert.Equal(t, "data.key1", diff.Entries[0].Key, "DiffEntry key should be key1")
	assert.Equal(t, SecretDiffEntryAdded, diff.Entries[1].Type, "DiffEntry type should be changed")
	assert.Equal(t, "data.key2", diff.Entries[1].Key, "DiffEntry key should be key1")
}

func TestSecretComparisonRemoveData(t *testing.T) {
	a := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {
			"key1": "value1",
			"key2": "value2",
		},
	}

	b := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {},
	}

	diff := CompareSecrets(a, b)
	diff.Print()
	assert.Equal(t, false, diff.Equal, "Secrets should not be equal")
	assert.Equal(t, SecretDiffEntryChanged, diff.Type, "Diff type should be changed")
	assert.Equal(t, 2, len(diff.Entries), "Diff should have 2 entry")
	assert.Equal(t, SecretDiffEntryRemoved, diff.Entries[0].Type, "DiffEntry type should be changed")
	assert.Equal(t, "data.key1", diff.Entries[0].Key, "DiffEntry key should be key1")
	assert.Equal(t, SecretDiffEntryRemoved, diff.Entries[1].Type, "DiffEntry type should be changed")
	assert.Equal(t, "data.key2", diff.Entries[1].Key, "DiffEntry key should be key1")
}

func TestSecretComparisonChangeData1(t *testing.T) {
	a := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {
			"key1": "value1",
			"key2": "value2",
		},
	}

	b := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {
			"key1": "newValue1",
			"key2": "value2",
		},
	}

	diff := CompareSecrets(a, b)
	diff.Print()
	assert.Equal(t, false, diff.Equal, "Secrets should not be equal")
	assert.Equal(t, SecretDiffEntryChanged, diff.Type, "Diff type should be changed")
	assert.Equal(t, 1, len(diff.Entries), "Diff should have 1 entry")
	assert.Equal(t, SecretDiffEntryChanged, diff.Entries[0].Type, "DiffEntry type should be changed")
	assert.Equal(t, "data.key1", diff.Entries[0].Key, "DiffEntry key should be key1")
}

func TestSecretComparisonChangeData2(t *testing.T) {
	a := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {
			"key2": "value2",
			"key1": "value1",
		},
	}

	b := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {
			"key1": "newValue1",
			"key2": "newValue2",
		},
	}

	diff := CompareSecrets(a, b)
	diff.Print()
	assert.Equal(t, false, diff.Equal, "Secrets should not be equal")
	assert.Equal(t, SecretDiffEntryChanged, diff.Type, "Diff type should be changed")
	assert.Equal(t, 2, len(diff.Entries), "Diff should have 2 entry")
	assert.Equal(t, SecretDiffEntryChanged, diff.Entries[0].Type, "DiffEntry type should be changed")
	assert.Equal(t, SecretDiffEntryChanged, diff.Entries[1].Type, "DiffEntry type should be changed")
	assert.Equal(t, "data.key1", diff.Entries[1].Key, "DiffEntry key should be key1")
	assert.Equal(t, "data.key2", diff.Entries[0].Key, "DiffEntry key should be key2")
}

func TestSecretComparisonChangeData3(t *testing.T) {
	a := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {
			"key1": "value1",
			"key2": "value2",
		},
	}

	b := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {
			"key1": "newValue1",
			"key3": "value3",
		},
	}

	diff := CompareSecrets(a, b)
	diff.Print()
	assert.Equal(t, false, diff.Equal, "Secrets should not be equal")
	assert.Equal(t, SecretDiffEntryChanged, diff.Type, "Diff type should be changed")
	assert.Equal(t, 3, len(diff.Entries), "Diff should have 3 entry")
	
	assert.Equal(t, SecretDiffEntryChanged, diff.Entries[0].Type, "DiffEntry type should be changed")
	assert.Equal(t, "data.key1", diff.Entries[0].Key, "DiffEntry key should be key1")

	assert.Equal(t, SecretDiffEntryRemoved, diff.Entries[1].Type, "DiffEntry type should be removed")
	assert.Equal(t, "data.key2", diff.Entries[1].Key, "DiffEntry key should be key2")

	assert.Equal(t, SecretDiffEntryAdded, diff.Entries[2].Type, "DiffEntry type should be added")
	assert.Equal(t, "data.key3", diff.Entries[2].Key, "DiffEntry key should be key3")
}

func TestSecretComparisonRemoveSecret(t *testing.T) {
	a := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {
			"key1": "value1",
			"key2": "value2",
		},
	}

	diff := CompareSecrets(a, nil)
	diff.Print()
	assert.Equal(t, false, diff.Equal, "Secrets should not be equal")
	assert.Equal(t, SecretDiffEntryRemoved, diff.Type, "Diff type should be removed")
	assert.Equal(t, 2, len(diff.Entries), "Diff should have 2 entry")
}

func TestSecretComparisonAddSecret(t *testing.T) {
	a := &Secret {
		Name: "myName",
		Namespace: "myNamespace",
		Type: "Opaque",
		Target: SecretFileTargetKubernetes,
		Data: map[string]string {
			"key1": "value1",
			"key2": "value2",
		},
	}

	diff := CompareSecrets(nil, a)
	diff.Print()
	assert.Equal(t, false, diff.Equal, "Secrets should not be equal")
	assert.Equal(t, SecretDiffEntryAdded, diff.Type, "Diff type should be added")
	assert.Equal(t, 2, len(diff.Entries), "Diff should have 2 entry")
}