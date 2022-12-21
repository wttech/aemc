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

sed -i '' 's/AEMC_VERSION:-"[^\"]*"/AEMC_VERSION:-"'"$VERSION"'"/g' init.sh
sed -i '' 's/AEMC_VERSION:-"[^\"]*"/AEMC_VERSION:-"'"$VERSION"'"/g' aemw
# shellcheck disable=SC2016
sed -i '' 's/aem\@v[^\`]*\`/aem@v'"$VERSION"\`'/g' README.MD

echo "Pushing version bump"
git commit -a -m "Release $VERSION_TAG"
git push

echo "Pushing release tag '$VERSION_TAG'"
git push origin "$VERSION_TAG"
