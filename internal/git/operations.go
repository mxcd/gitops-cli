package git

import (
	git2go "github.com/libgit2/git2go/v34"
)

type FetchOptionsUsage string

const (
	FetchOptionUsageDefault FetchOptionsUsage = "default"
)

func (c *Connection) getFetchOptions(usage FetchOptionsUsage) git2go.FetchOptions {

	fetchOptions := git2go.FetchOptions{}

	switch usage {
	case FetchOptionUsageDefault:
		fetchOptions.RemoteCallbacks = git2go.RemoteCallbacks{
			CredentialsCallback: c.credentialsCallback,
		}
	}

	if c.Options.IgnoreSslHostKey {
		fetchOptions.RemoteCallbacks.CertificateCheckCallback = func(cert *git2go.Certificate, valid bool, hostname string) error {
			return nil
		}
	}

	return fetchOptions
}
