#!/usr/bin/env sh

AEM_WRAPPER="aemw"

if [ -f "$AEM_WRAPPER" ]; then
  echo "The project contains already AEM Compose!"
  exit 1
fi

VERSION=${AEM_CLI_VERSION:-"0.16.3"}
SOURCE_URL="https://raw.githubusercontent.com/wttech/aemc/v${VERSION}/project/cloud"

AEM_DIR="aem"
HOME_DIR="${AEM_DIR}/home"
SCRIPT_DIR="${AEM_DIR}/script"
LIB_DIR="${HOME_DIR}/lib"

DEFAULT_DIR="${AEM_DIR}/default"
DEFAULT_CONFIG_DIR="${DEFAULT_DIR}/etc"

echo ""
echo "Downloading AEM Compose Files"
echo ""

mkdir -p "${HOME_DIR}"

mkdir -p "${SCRIPT_DIR}"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/deploy.sh" -o "${SCRIPT_DIR}/deploy.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/destroy.sh" -o "${SCRIPT_DIR}/destroy.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/down.sh" -o "${SCRIPT_DIR}/down.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/provision.sh" -o "${SCRIPT_DIR}/provision.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/resetup.sh" -o "${SCRIPT_DIR}/resetup.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/restart.sh" -o "${SCRIPT_DIR}/restart.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/setup.sh" -o "${SCRIPT_DIR}/setup.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/undeploy.sh" -o "${SCRIPT_DIR}/undeploy.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/up.sh" -o "${SCRIPT_DIR}/up.sh"

mkdir -p "${DEFAULT_CONFIG_DIR}"
curl -s "${SOURCE_URL}/${DEFAULT_CONFIG_DIR}/aem.yml" -o "${DEFAULT_CONFIG_DIR}/aem.yml"

mkdir -p "${AEM_DIR}"
curl -s "${SOURCE_URL}/${AEM_WRAPPER}" -o "${AEM_WRAPPER}"

mkdir -p "${LIB_DIR}"

echo "Downloading & Testing AEM Compose CLI"
echo ""

chmod +x "${AEM_WRAPPER}"
sh ${AEM_WRAPPER} version

echo "Success! Now initialize AEM Compose by running the command:"
echo ""

echo "sh ${AEM_WRAPPER} init"
