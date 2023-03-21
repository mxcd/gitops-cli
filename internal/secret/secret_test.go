package secret

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadSecret1(t *testing.T) {
	f := "../../test/test.secret.enc.yml"
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
	f := "../../test/implicit-name.secret.enc.yml"
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