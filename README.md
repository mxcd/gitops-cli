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

## Usage

```
NAME:
   gitpos - GitOps CLI

USAGE:
   gitpos [global options] command [command options] [arguments...]

COMMANDS:
   secrets, s  GitOps managed secrets
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --root-dir value              root directory of the git repository [$GITOPS_ROOT_DIR]
   --kubeconfig value, -k value  kubeconfig file to use for connecting to the Kubernetes cluster [$KUBECONFIG, $GITOPS_KUBECONFIG]
   --verbose, -v                 debug output (default: false) [$GITOPS_VERBOSE]
   --very-verbose, --vv          trace output (default: false) [$GITOPS_VERY_VERBOSE]
   --cleartext                   print secrets in cleartext to the console (default: false) [$GITOPS_CLEARTEXT]
   --print                       print secrets to the console (default: false) [$GITOPS_PRINT]
   --help, -h                    show help
```

### Planning secret application to a cluster

**NOTE:** It is expected, that the cluster's KUBECONFIG is already set up. Alternatively, the `--kubeconfig` flag can be used.

```bash
gitops secrets plan kubernetes
```

Or in short

```bash
gitops s p k8s
```

Example output:

```
__             _ __
\ \     ____ _(_) /_____  ____  _____
 \ \   / __ `/ / __/ __ \/ __ \/ ___/
 / /  / /_/ / / /_/ /_/ / /_/ (__  )
/_/   \__, /_/\__/\____/ .___/____/
     /____/           /_/

GitOps CLI computed the following changes for your cluster:
-------------------------------------------------------

default/my-config-map :  add
  + data.loremIpsum: **************************************************
  + data.someConfigMapKey: **************
---
default/implicit-name :  unchanged
---
default/my-secret-name :  unchanged
---
default/database-credentials :  change
  ~ data.bar: ** => **

-------------------------------------------------------

use gitops secrets apply kubernetes to apply these changes to your cluster
```

### Applying secrets to a cluster

**NOTE:** It is expected, that the cluster's KUBECONFIG is already set up. Alternatively, the `--kubeconfig` flag can be used.

```bash
gitops secrets apply kubernetes
```

Or in short

```bash
gitops s a k8s
```

The user will be prompted to confirm the changes before they are applied to the cluster. The prompt can be bypassed by using the `--auto-approve` flag.  
Example output:

```
__             _ __
\ \     ____ _(_) /_____  ____  _____
 \ \   / __ `/ / __/ __ \/ __ \/ ___/
 / /  / /_/ / / /_/ /_/ / /_/ (__  )
/_/   \__, /_/\__/\____/ .___/____/
     /____/           /_/

GitOps CLI computed the following changes for your cluster:
-------------------------------------------------------

default/my-config-map :  add
  + data.someConfigMapKey: **************
  + data.loremIpsum: **************************************************
---
default/implicit-name :  unchanged
---
default/my-secret-name :  unchanged
---
default/database-credentials :  change
  ~ data.bar: ** => *****

-------------------------------------------------------

GitOps CLI will apply these changes to your Kubernetes cluster.
Only 'yes' will be accepted to approve.
Apply changes above: yes
GitOps CLI will now execute the changes for your cluster:
-------------------------------------------------------

default / my-config-map  created
default / database-credentials  updated

-------------------------------------------------------

All changes applied.
```

Redacted secrets (`*********`) can be displayed in cleartext by using the `--cleartext` flag.  
To print all loaded secrets to the console, use the `--print` flag.

## Installation

### MacOS

Install using homebrew:

```bash
brew tap mxcd/gitops
brew install gitops
```

## Features

### GitOps secret management

The GitOps CLI can handle secrets in a GitOps way. Either by injecting them directly
as K8s secrets or by sending them to a vault instance for safekeeping. Either way,
the secrets are stored in a Git repository and secured using SOPS.

#### Secret storage

Secrets are stored in any directory of your git repository. The GitOps CLI will pick
up any file that ends with `*.gitops.secret.enc.y[a]ml` except for `values.gitops.secret.enc.y[a]ml` (see [Secrets Templating](#secrets-templating))
The secret files must be encrypted using SOPS.

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
# target of the secret
targetType: < k8s | vault >

# name of the secret
name: <my-secret-name>
# optional namespace of the secret (default: default)
namespace: <my-namespace>
# type of the secret (default: Opaque)
# only for k8s secrets: ConfigMap or any of the following: https://kubernetes.io/docs/concepts/configuration/secret/#secret-types
type: < ConfigMap | Opaque | ... >
# data of the secret as kv pairs
data:
  <key1>: <value1>
  <key2>: <value2>
```

##### Case 1: Secret for K8s

```yaml
# targetType of the secret
targetType: k8s
# name of the secret
name: my-secret-name
# optional namespace of the secret (default: default)
namespace: my-namespace
# type of the secret (default: Opaque)
type: Opaque
# data of the secret as kv pairs
data:
  key: value
```

If the name is not given in the file, the name will be inferred from the filename. The file extension `.gitops.secret.enc.y[a]ml` will be removed.

```yaml
my-secret-name.gitops.secret.enc.yaml
# will be applied as
name: my-secret-name
```

This implies, that the filename must be a valid K8s secret name.

##### Case 2: Secret for Vault

**NOTE:** Vault secrets are still WIP

```yaml
# target of the secret
targetType: vault
# name of the secret - will be used as path in vault
name: /my/secret/name
# data of the secret as kv pairs
data:
  key: value
```

#### Secrets Templating

It is possible to use Go templates in the secret files. The values will originate from sops-encrypted `values.gitops.secret.enc.y[a]ml` files.  
Values files can be located anywhere in the repository. The GitOps CLI will pick up all files that are located on the direct path towards the respective secret file.  
Values files closer to the secret file will have higher precedence. Any object structure is allowed to be used in a values file.

Example: 

```yaml
# /foo/bar/dev/values.gitops.secret.enc.yaml
environment: dev
database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
```

```yaml
# /foo/bar/dev/values.gitops.secret.enc.yaml
targetType: k8s
# name of the secret - will be used as path in vault
name: my-service
namespace: "{{ .Values.environment }}"
data:
  application.properties: |-
    service.environment={{ .Values.environment }}
    spring.datasource.url=jdbc:postgresql://{{ .Values.database.host }}:{{ .Values.database.port }}/mydb
    spring.datasource.username={{ .Values.database.user }}
    spring.datasource.password={{ .Values.database.password }}
```
**NOTE** that the template string (`{{ .Values.someValue }}`) must be enclosed in quotes for sops to work properly. In the above example, the entire `application.properties` data value is considered as a string and thus does not need further quoting.

#### Multi-cluster support
It is possible to address multiple clusters with a single GitOps repository.  
To add a new cluster to the GitOps state use
```
gitops clusters add <cluster-name> <kubeconfig-path>
```
The kubeconfig file can either be a plain text file or a sops-encrypted file. If the file is encrypted, it must adhere to the following naming convention to be decrypted properly: 
```
*.kubeconfig.secret.enc.ya?ml
```
  
To inspect the currently configured clusters use
```
gitops clusters list
```
```
__             _ __
\ \     ____ _(_) /_____  ____  _____
 \ \   / __ `/ / __/ __ \/ __ \/ ___/
 / /  / /_/ / / /_/ /_/ / /_/ (__  )
/_/   \__, /_/\__/\____/ .___/____/
     /____/           /_/

dev  =>  kubeconfigs/dev.kubeconfig.secret.enc.yml
int  =>  kubeconfigs/int.kubeconfig.secret.enc.yml
prod  =>  kubeconfigs/prod.kubeconfig.secret.enc.yml
```
  
To remove a cluster from the GitOps state use
```
gitops clusters remove <cluster-name>
```

To check connectivity with the configured clusters use
```
gitops clusters test
```
```
__             _ __
\ \     ____ _(_) /_____  ____  _____
 \ \   / __ `/ / __/ __ \/ __ \/ ___/
 / /  / /_/ / / /_/ /_/ / /_/ (__  )
/_/   \__, /_/\__/\____/ .___/____/
     /____/           /_/

Cluster: __default (v1.25.3)    Connected: true
Cluster: dev (v1.24.8)          Connected: true
Cluster: int (v1.24.8)          Connected: true
Cluster: prod (v1.24.8)         Connected: true
```
  
A secret can be configured to be applied to a specific cluster using the `target` attribute in the secret file. The default value is the `__default` cluster which is inferred from the `KUBECONFIG` environment variable or the default kubeconfig file. The `target` attribute can also be set using a templating variable so that all secrets under a certain directory will be applied to a specific cluster.
```yaml
# /foo/bar/dev/values.gitops.secret.enc.yaml
target: dev
```
```yaml
# /foo/bar/dev/myservice/service.gitops.secret.enc.yaml
targetType: k8s
target: "{{ .Values.target }}" # will be replaced with "dev"
name: my-service
data:
  key: value
```
  
By default, secrets will be applied to all configured clusters. This can be limited by giving the cluster as an argument:
```
gitops secrets plan kubernetes dev
__             _ __
\ \     ____ _(_) /_____  ____  _____
 \ \   / __ `/ / __/ __ \/ __ \/ ___/
 / /  / /_/ / / /_/ /_/ / /_/ (__  )
/_/   \__, /_/\__/\____/ .___/____/
     /____/           /_/

Limiting to cluster dev

[Loading local secrets] 100% |██████████████████████████████████████████████████| (63/63) 


No changes to apply.
```

#### Directory limiter
It is possible to restrict the secrets input to a specific directory to speed up loading and decryption of secrets. This can be done by providing the `--dir` flag:
```
gitops secrets --dir application/dev plan kubernetes

__             _ __
\ \     ____ _(_) /_____  ____  _____
 \ \   / __ `/ / __/ __ \/ __ \/ ___/
 / /  / /_/ / / /_/ /_/ / /_/ (__  )
/_/   \__, /_/\__/\____/ .___/____/
     /____/           /_/

Limiting to directory applications/dev

[Loading local secrets] 100% |██████████████████████████████████████████████████| (1/1) 

No changes to apply.
```
**NOTE** that the directory path must be relative to the repository root and that only forward slashes (`/`) are supported.
## Repository

### After the first clone

#### Pre-commit

Please make sure to follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) when committing to this repository.  
To make one's life easier, a pre-commit config is provided that can be installed with the following command:

```bash
pre-commit install --hook-type commit-msg
```

## Repo Server
The repo server allows to run the GitOps CLI in a server mode. 
This is especially useful if many concurrent patches are to be expected (e.g. due to matrix builds of multiple services). 
The server will clone the defined GitOps cluster repository and apply patches when instructed via the REST API.

### Running the server
The most basic configuration uses https and basic auth:
```bash
docker run --platform linux/amd64 --rm -it -p 8080:8080 \
  -e GITOPS_REPOSITORY="https://github.com/<owner>/<repo>" \
  -e GITOPS_REPOSITORY_BASICAUTH="<username>:<personal-access-token>" \
  -e API_KEYS="<super-secret-api-key>" \
  ghcr.io/mxcd/gitops/repo-server:2.2.3
```

`API_KEYS` is a comma separated list of API keys that are used to authenticate the requests.

### Using the server
The server exposes a REST API that can be used to apply patches to the GitOps repository.
The API is available at `PUT <protocol>://<host>:<port>/api/v1/patch`  

```bash
# env vars used for example curl BUT also valid for CLI usage
export GITOPS_REPOSITORY_SERVER="<protocol>://<host>[:<port>]/api/v1"
export GITOPS_REPOSITORY_SERVER_API_KEY="<super-secret-api-key>"
```

```bash
curl --request PUT \
  --url $GITOPS_REPOSITORY_SERVER/patch \
  --header 'content-type: application/json' \
  --header 'X-API-Key: $GITOPS_REPOSITORY_SERVER_API_KEY'
  --data '{
  "filePath": "applications/dev/service-foo/values.yaml",
  "patches": [
    {
      "selector": ".service.image.tag",
      "value": "v42.0.1"
    }
  ]
}'
```

```bash
# Equivalent CLI command
# Evaluates the previously exported env vars GITOPS_REPOSITORY_SERVER and GITOPS_REPOSITORY_SERVER_API_KEY
gitops patch applications/dev/service-foo/values.yaml .service.image.tag v42.0.1
```