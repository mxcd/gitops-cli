#!/bin/bash

USERNAME=""
TOKEN=""
VERSION=""
DRYRUN="false"
VERBOSE="false"

TARGET_REPOSITORY="github.com/mxcd/homebrew-gitops"

while getopts 't:v:d' flag; do
  case "${flag}" in
    u) USERNAME="${OPTARG}" ;;
    t) TOKEN="${OPTARG}" ;;
    v) VERSION="${OPTARG}" ;;
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

log "Publishing version $VERSION"

git clone https://$USERNAME:$TOKEN@$TARGET_REPOSITORY.git