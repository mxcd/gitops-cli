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

# cleanup
rm -rf $BASE/hack/soft-serve/ssh-key
rm -rf $BASE/hack/soft-serve/ssh-key.pub
rm -rf $BASE/hack/soft-serve/data
rm -rf $BASE/hack/soft-serve/gitops-test

remove_ssh_key() {
  ssh-add -d $BASE/hack/soft-serve/ssh-key.pub  
}

trap remove_ssh_key INT

# generate ssh key
ssh-keygen -t ed25519 -N '' -f $BASE/hack/soft-serve/ssh-key

# set config vars
export SOFT_SERVE_INITIAL_ADMIN_KEYS=$(cat $BASE/hack/soft-serve/ssh-key.pub)

docker compose -f $BASE/hack/soft-serve/docker-compose.yml up -d

sleep 1

ssh-add $BASE/hack/soft-serve/ssh-key
echo "Creating repo"
ssh -p 23231 -o StrictHostKeychecking=no -i $BASE/hack/soft-serve/ssh-key localhost repo create gitops-test

echo "\nCloning repo"
export GIT_SSH_COMMAND="ssh -o StrictHostKeychecking=no -i $BASE/hack/soft-serve/ssh-key"
git clone ssh://localhost:23231/gitops-test.git $BASE/hack/soft-serve/gitops-test
mkdir -p $BASE/hack/soft-serve/gitops-test/applications/dev/service-test
cp $BASE/hack/soft-serve/fixtures/values.yaml $BASE/hack/soft-serve/gitops-test/applications/dev/service-test/values.yaml
cd $BASE/hack/soft-serve/gitops-test
git config user.name "Soft-Serve"
git config user.email "soft-serve@localhost"
git checkout -b main
git add .
git commit -m "feat: add service-test application"
git push origin main

cd $BASE/hack/soft-serve