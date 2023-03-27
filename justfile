set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

# This list of available targets
default:
    @just --summary

# pushes all changes to the main branch
push +COMMIT_MESSAGE:
  git add .
  git commit -m "{{COMMIT_MESSAGE}}"
  git pull origin $(git rev-parse --abbrev-ref HEAD)
  git push origin $(git rev-parse --abbrev-ref HEAD)