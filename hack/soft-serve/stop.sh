#!/bin/bash

BASE="$(git rev-parse --show-toplevel)"
[[ $? -eq 0 ]] || {
  echo 'Run this script from inside the repository (cannot determine toplevel directory)'
  exit 1
}

log() {
  if [[ $VERBOSE == 'true' ]]; then
    echo $1
  fi
}

docker compose -f $BASE/hack/soft-serve/docker-compose.yml down

# cleanup
rm -rf $BASE/hack/soft-serve/ssh-key
rm -rf $BASE/hack/soft-serve/ssh-key.pub
rm -rf $BASE/hack/soft-serve/data
rm -rf $BASE/hack/soft-serve/gitops-test
