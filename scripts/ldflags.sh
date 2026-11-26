#!/usr/bin/env sh
# Prints Go linker flags that stamp the CLI binary with its Git version, commit and commit date.

set -e

GIT_COMMIT=$(git rev-parse HEAD)
GIT_COMMIT_DATE=$(git log -1 --date=format:'%Y-%m-%dT%H:%M:%S' --format=%cd)
GIT_VERSION=$(git describe --tags 2>/dev/null || echo "$GIT_COMMIT")

echo "-s -w -X main.appVersion=$GIT_VERSION -X main.appCommit=$GIT_COMMIT -X main.appCommitDate=$GIT_COMMIT_DATE"
