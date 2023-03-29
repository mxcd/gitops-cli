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
	entry := diff.GetEntry("target")
	assert.NotNil(t, entry, "Diff should have an entry for target")
	assert.Equal(t, SecretDiffEntryChanged, entry.Type, "DiffEntry type should be changed")
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
	entry := diff.GetEntry("type")
	assert.NotNil(t, entry, "Diff should have an entry for type")
	assert.Equal(t, SecretDiffEntryChanged, entry.Type, "DiffEntry type should be changed")
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
	
	entry := diff.GetEntry("name")
	assert.NotNil(t, entry, "Diff should have an entry for name")
	assert.Equal(t, SecretDiffEntryChanged, entry.Type, "DiffEntry type should be changed")
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
	
	entry := diff.GetEntry("namespace")
	assert.NotNil(t, entry, "Diff should have an entry for namespace")
	assert.Equal(t, SecretDiffEntryChanged, entry.Type, "DiffEntry type should be changed")
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
	
	entry1 := diff.GetEntry("data.key1")
	assert.NotNil(t, entry1, "Diff should have an entry for data.key1")
	assert.Equal(t, SecretDiffEntryAdded, entry1.Type, "DiffEntry type should be added")
	
	entry2 := diff.GetEntry("data.key2")
	assert.NotNil(t, entry2, "Diff should have an entry for data.key2")
	assert.Equal(t, SecretDiffEntryAdded, entry2.Type, "DiffEntry type should be added")
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

	entry1 := diff.GetEntry("data.key1")
	assert.NotNil(t, entry1, "Diff should have an entry for data.key1")
	assert.Equal(t, SecretDiffEntryRemoved, entry1.Type, "DiffEntry type should be removed")
	
	entry2 := diff.GetEntry("data.key2")
	assert.NotNil(t, entry2, "Diff should have an entry for data.key2")
	assert.Equal(t, SecretDiffEntryRemoved, entry2.Type, "DiffEntry type should be removed")
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

	entry1 := diff.GetEntry("data.key1")
	assert.NotNil(t, entry1, "Diff should have an entry for data.key1")
	assert.Equal(t, SecretDiffEntryChanged, entry1.Type, "DiffEntry type should be changed")
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

	entry1 := diff.GetEntry("data.key1")
	assert.NotNil(t, entry1, "Diff should have an entry for data.key1")
	assert.Equal(t, SecretDiffEntryChanged, entry1.Type, "DiffEntry type should be changed")
	
	entry2 := diff.GetEntry("data.key2")
	assert.NotNil(t, entry2, "Diff should have an entry for data.key2")
	assert.Equal(t, SecretDiffEntryChanged, entry2.Type, "DiffEntry type should be changed")
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
	
	entry1 := diff.GetEntry("data.key1")
	assert.NotNil(t, entry1, "Diff should have an entry for data.key1")
	assert.Equal(t, SecretDiffEntryChanged, entry1.Type, "DiffEntry type should be added")
	
	entry2 := diff.GetEntry("data.key2")
	assert.NotNil(t, entry2, "Diff should have an entry for data.key2")
	assert.Equal(t, SecretDiffEntryRemoved, entry2.Type, "DiffEntry type should be removed")

	entry3 := diff.GetEntry("data.key3")
	assert.NotNil(t, entry3, "Diff should have an entry for data.key3")
	assert.Equal(t, SecretDiffEntryAdded, entry3.Type, "DiffEntry type should be added")
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

	entry1 := diff.GetEntry("data.key1")
	assert.NotNil(t, entry1, "Diff should have an entry for data.key1")
	assert.Equal(t, SecretDiffEntryRemoved, entry1.Type, "DiffEntry type should be removed")

	entry2 := diff.GetEntry("data.key2")
	assert.NotNil(t, entry1, "Diff should have an entry for data.key2")
	assert.Equal(t, SecretDiffEntryRemoved, entry2.Type, "DiffEntry type should be removed")
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

	entry1 := diff.GetEntry("data.key1")
	assert.NotNil(t, entry1, "Diff should have an entry for data.key1")
	assert.Equal(t, SecretDiffEntryAdded, entry1.Type, "DiffEntry type should be added")

	entry2 := diff.GetEntry("data.key2")
	assert.NotNil(t, entry1, "Diff should have an entry for data.key2")
	assert.Equal(t, SecretDiffEntryAdded, entry2.Type, "DiffEntry type should be added")
}