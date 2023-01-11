#!/usr/bin/env sh

VERSION=$1
VERSION_TAG="v$VERSION"

if [ -z "$VERSION" ]
then
      echo "Release version is not specified!"
      exit 1
fi

GIT_STAT=$(git diff --stat)

if [ "$GIT_STAT" != '' ]; then
  echo "Unable to release. Uncommited changes detected!"
  exit 1
fi

echo "Releasing $VERSION_TAG"

echo "Bumping version in files"

README_FILE="README.MD"
PROJECT_INIT_SCRIPT="project/init.sh"
PROJECT_WRAPPER_SCRIPT="project/aemw"

# <https://stackoverflow.com/a/57766728>
if [ "$(uname)" = "Darwin" ]; then
  sed -i '' 's/AEMC_VERSION:-"[^\"]*"/AEMC_VERSION:-"'"$VERSION"'"/g' "$PROJECT_INIT_SCRIPT"
  sed -i '' 's/AEMC_VERSION:-"[^\"]*"/AEMC_VERSION:-"'"$VERSION"'"/g' "$PROJECT_WRAPPER_SCRIPT"
  # shellcheck disable=SC2016
  sed -i '' 's/aem\@v[^\`]*\`/aem@v'"$VERSION"\`'/g' "$README_FILE"
else
    sed -i 's/AEMC_VERSION:-"[^\"]*"/AEMC_VERSION:-"'"$VERSION"'"/g' "$PROJECT_INIT_SCRIPT"
    sed -i 's/AEMC_VERSION:-"[^\"]*"/AEMC_VERSION:-"'"$VERSION"'"/g' "$PROJECT_WRAPPER_SCRIPT"
    # shellcheck disable=SC2016
    sed -i 's/aem\@v[^\`]*\`/aem@v'"$VERSION"\`'/g' "$README_FILE"
fi

echo "Pushing version bump"
git commit -a -m "Release $VERSION_TAG"
git push

echo "Pushing release tag '$VERSION_TAG'"
git tag "$VERSION_TAG"
git push origin "$VERSION_TAG"
