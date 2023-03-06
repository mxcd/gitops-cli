```
__             _ __
\ \     ____ _(_) /_____  ____  _____
 \ \   / __ `/ / __/ __ \/ __ \/ ___/
 / /  / /_/ / / /_/ /_/ / /_/ (__  )
/_/   \__, /_/\__/\____/ .___/____/
     /____/           /_/
```

[//]: # "https://patorjk.com/software/taag/#p=display&f=Slant&t=%3E%20gitops"

## GitOps CLI

CLI tool for performing GitOps operations

## Features

### GitOps secret management

The GitOps CLI can handle secrets in a GitOps way. Either by injecting them directly
as K8s secrets or by sending them to a vault instance for safekeeping. Either way,
the secrets are stored in a Git repository and secured using SOPS.

#### Secret storage

Secrets are stored in any directory of your git repository. The GitOps CLI will pick
up any file that ends with `*.secret.enc.yml` or `*.secret.enc.yaml`. The secret files
must be encrypted using SOPS.

**NOTE:** Secrets MUST NEVER be committed into version control unencrypted.
Therefore, it is very much encouraged to add the following lines to your `.gitignore` file:

```gitignore
*.secret.yaml
*.secret.yml
*.secret.env
```

Make sure to follow a strict naming convention for your secret files, in order to keep them matching those patterns.

#### Secrets file format

The secrets files must follow the following format:

```yaml
# name of the secret
name: secret-name
# namespace of the secret (only applicable for K8s)
namespace: secret-namespace
# data of the secret
# only KV pairs are supported
data:
  key: value
```

To ensure intercompatibility with K8s and vault, the following rules apply:

If the name is not given in the file, the name will be inferred from the filename.

```yaml
my-secret-name.secret.enc.yaml
# will be applied as
# K8s:
name: my-secret-name
# Vault path:
/my/secret/name
```

When applying to vault, dashes in the file name will be converted to slashes:

```yaml
# filename:
my-secret-name.secret.enc.yaml
# Vault path:
/my/secret/name
```

When applying to K8s, slashes in the name will be converted to dashes:

```yaml
name: my/secret/name
# will be applied as
name: my-secret-name
```


## Repository

### After the first clone

#### Pre-commit

Please make sure to follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) when committing to this repository.  
To make one's life easier, a pre-commit config is provided that can be installed with the following command:

```bash
pre-commit install --hook-type commit-msg
```