#!/usr/bin/env sh

# Installs AEM Compose into the current project, run by mise (default) or Task:
#   curl https://raw.githubusercontent.com/wttech/aemc/main/project-install.sh | sh
#   curl https://raw.githubusercontent.com/wttech/aemc/main/project-install.sh | AEM_RUNNER=task sh

RUNNER="${AEM_RUNNER:-mise}"
SOURCE_URL="https://raw.githubusercontent.com/wttech/aemc/main/pkg/project/runner/${RUNNER}/common"

if [ -f "mise.toml" ] || [ -f "aemw" ]; then
  echo ""
  echo "The project already contains AEM Compose!"
  exit 0
fi

download() {
  mkdir -p "$(dirname "$2")"
  if ! curl -fsSL "${SOURCE_URL}/$1" -o "$2"; then
    echo "Error! Cannot download '${SOURCE_URL}/$1'"
    exit 1
  fi
}

echo ""
echo "Downloading & Testing AEM Compose CLI (runner: ${RUNNER})"
echo ""

case "${RUNNER}" in
  mise)
    download "bin/mise" "bin/mise"
    download "bin/mise.cmd" "bin/mise.cmd"
    chmod +x "bin/mise"
    AEM="./bin/mise exec github:wttech/aemc@latest -- aem"
    ;;
  task)
    download "aemw" "aemw"
    chmod +x "aemw"
    AEM="sh aemw"
    ;;
  *)
    echo "Error! Runner '${RUNNER}' is not supported, use 'mise' or 'task'"
    exit 1
    ;;
esac

$AEM version || exit 1

echo ""
echo "Success! Now scaffold the AEM Compose files in the project by running command below:"
echo ""
echo "${AEM} project scaffold --runner ${RUNNER}"
