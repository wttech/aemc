#!/usr/bin/env sh
# Releases a new version: bumps it in files, pushes the commit and then the tag that triggers the 'Release Perform' workflow.
# Usage: sh scripts/release.sh <version>   (without the leading 'v', e.g. 2.0.20)

set -e

VERSION="$1"
VERSION_TAG="v$VERSION"

VERSION_CURRENT_TAG=$(git describe --tags --abbrev=0 || echo "v0.0.0")
VERSION_CURRENT_TAG="${VERSION_CURRENT_TAG#v}" # remove leading 'v', preserve leading zeros

if [ -z "$VERSION" ]; then
    echo "Release version is not specified!"
    echo "Last released: ${VERSION_CURRENT_TAG}"
    exit 1
fi

GIT_STAT=$(git diff --stat || true)

if [ "$GIT_STAT" != '' ]; then
    echo "Unable to release. Uncommitted changes detected!"
    exit 1
fi

echo ""
echo "Releasing $VERSION_TAG"
echo ""

echo "Bumping version in files"
bump_version() {
    local file="$1"
    if [ "$(uname)" = "Darwin" ]; then
        sed -i '' 's/AEM_CLI_VERSION:-"[^\"]*"/AEM_CLI_VERSION:-"'"$VERSION"'"/g' "$file"
        # shellcheck disable=SC2016
        sed -i '' 's/aem\@v[^\`]*\`/aem@v'"$VERSION"\`'/g' "$file"
    else
        sed -i 's/AEM_CLI_VERSION:-"[^\"]*"/AEM_CLI_VERSION:-"'"$VERSION"'"/g' "$file"
        # shellcheck disable=SC2016
        sed -i 's/aem\@v[^\`]*\`/aem@v'"$VERSION"\`'/g' "$file"
    fi
}
bump_version "README.md"
bump_version "pkg/project/common/aemw"

echo "Pushing version bump"
git commit -a -m "Release $VERSION_TAG" || echo "No changes to commit"
git push || echo "Nothing to push"

echo "Pushing release tag '$VERSION_TAG'"
git tag "$VERSION_TAG"
git push origin "$VERSION_TAG" || echo "Tag already exists or cannot push"
