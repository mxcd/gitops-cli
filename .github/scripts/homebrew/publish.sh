#!/bin/bash

USERNAME=""
TOKEN=""
VERSION=""
SHA=""
DRYRUN="false"
VERBOSE="false"

BASE="$(git rev-parse --show-toplevel)"
[[ $? -eq 0 ]] || {
  echo 'Run this script from inside the repository (cannot determine toplevel directory)'
  exit 1
}

while getopts 'u:t:v:s:d' flag; do
  case "${flag}" in
    u) USERNAME="${OPTARG}" ;;
    t) TOKEN="${OPTARG}" ;;
    v) VERSION="${OPTARG}" ;;
    s) SHA="${OPTARG}" ;;
    d) DRYRUN='true' ;;
    *) echo "Unexpected option ${flag}"; exit 1 ;;
  esac
done

log() {
  if [[ $VERBOSE == 'true' ]]; then
    echo $1
  fi
}

if [[ $USERNAME == "" ]]; then
  echo "Missing username"
  exit 1
fi

if [[ $TOKEN == "" ]]; then
  echo "Missing token"
  exit 1
fi

if [[ $VERSION == "" ]]; then
  echo "Missing version"
  exit 1
fi

if [[ $SHA == "" ]]; then
  echo "Missing SHA"
  exit 1
fi

log "Publishing version $VERSION with SHA $SHA"

cd $BASE
git clone https://$USERNAME:$TOKEN@github.com/homebrew-gitops.git 
export RELEASE_SHA256="$SHA"
export RELEASE_VERSION="$VERSION"

cat $BASE/.github/scripts/homebrew/gitops.rb | envsubst > $BASE/homebrew-gitops/Formula/gitops.rb

cd $BASE/homebrew-gitops
git add Formula/gitops.rb
git commit -m "feat: updating to version $VERSION"

if [[ $DRYRUN == 'true' ]]; then
  echo "Dry run, not pushing"
else
  git push
fi

cd $BASE