#!/usr/bin/env sh

VERSION=${AEMC_VERSION:-"0.7.1"}
SOURCE_URL="https://raw.githubusercontent.com/wttech/aemc/v${VERSION}/project"

AEM_WRAPPER="aemw"
AEM_DIR="aem"
SCRIPT_DIR="${AEM_DIR}/script"
HOME_DIR="${AEM_DIR}/home"
LIB_DIR="${HOME_DIR}/lib"
CONFIG_FILE="${HOME_DIR}/aem.yml"
SETUP_FILE="${SCRIPT_DIR}/setup.sh"

if [ -f "$AEM_WRAPPER" ]; then
  echo "The project contains already AEM Compose!"
  exit 1
fi

echo "Downloading AEM Compose Files"
echo ""

mkdir -p "$SCRIPT_DIR" "$HOME_DIR"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/destroy.sh" -o "${SCRIPT_DIR}/destroy.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/down.sh" -o "${SCRIPT_DIR}/down.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/resetup.sh" -o "${SCRIPT_DIR}/resetup.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/restart.sh" -o "${SCRIPT_DIR}/restart.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/setup.sh" -o "${SCRIPT_DIR}/setup.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/up.sh" -o "${SCRIPT_DIR}/up.sh"
curl -s "${SOURCE_URL}/${AEM_DIR}/api.sh" -o "${AEM_DIR}/api.sh"
curl -s "${SOURCE_URL}/${AEM_WRAPPER}" -o "${AEM_WRAPPER}"

echo "Downloading & Running AEM Compose CLI"
echo ""

chmod +x "${AEM_WRAPPER}"
sh ${AEM_WRAPPER} version

echo "Scaffolding AEM Compose configuration file"
echo ""

./${AEM_WRAPPER} config init

echo "Creating AEM Compose directories"
echo ""

mkdir -p "$LIB_DIR"

echo "Initialized AEM Compose"
echo ""

echo "The next step is providing AEM files (JAR or SDK ZIP, license) to directory '${LIB_DIR}'"
echo "Alternatively, instruct the tool where these files are located by adjusting properties: 'dist_file', 'license_file' in configuration file '${CONFIG_FILE}'"
echo "Later on, remember to customise AEM instance setup in provisioning file '${SETUP_FILE}' for service pack installation, application build, etc."
echo "To avoid problems with IDE performance, make sure to exclude from indexing the directory '${HOME_DIR}'"
echo "Finally, use control scripts to manage AEM instances:"
echo ""

echo "sh aemw [setup|resetup|up|down|restart]"

echo ""
echo "It is also possible to run individual AEM Compose CLI commands separately."
echo "Discover available commands by running:"
echo ""

echo "sh aemw --help"
