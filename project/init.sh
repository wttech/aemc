#!/usr/bin/env sh

VERSION=${AEMC_VERSION:-"0.12.1"}
SOURCE_URL="https://raw.githubusercontent.com/wttech/aemc/v${VERSION}/project"

AEM_WRAPPER="aemw"

AEM_DIR="aem"
SCRIPT_DIR="${AEM_DIR}/script"
HOME_DIR="${AEM_DIR}/home"
DEFAULT_DIR="${AEM_DIR}/default"
DEFAULT_CONFIG_DIR="${DEFAULT_DIR}/etc"
LIB_DIR="${HOME_DIR}/lib"

if [ -f "$AEM_WRAPPER" ]; then
  echo "The project contains already AEM Compose!"
  exit 1
fi

echo "Downloading AEM Compose Files"
echo ""

mkdir -p "${SCRIPT_DIR}" "${HOME_DIR}" "${DEFAULT_CONFIG_DIR}" "${LIB_DIR}"

curl -s "${SOURCE_URL}/${DEFAULT_CONFIG_DIR}/aem.yml" -o "${DEFAULT_CONFIG_DIR}/aem.yml"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/deploy.sh" -o "${SCRIPT_DIR}/deploy.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/destroy.sh" -o "${SCRIPT_DIR}/destroy.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/down.sh" -o "${SCRIPT_DIR}/down.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/resetup.sh" -o "${SCRIPT_DIR}/resetup.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/restart.sh" -o "${SCRIPT_DIR}/restart.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/setup.sh" -o "${SCRIPT_DIR}/setup.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/up.sh" -o "${SCRIPT_DIR}/up.sh"
curl -s "${SOURCE_URL}/${AEM_DIR}/api.sh" -o "${AEM_DIR}/api.sh"
curl -s "${SOURCE_URL}/${AEM_WRAPPER}" -o "${AEM_WRAPPER}"

echo "Downloading & Testing AEM Compose CLI"
echo ""

chmod +x "${AEM_WRAPPER}"
sh ${AEM_WRAPPER} version

echo "Success! Now initialize AEM Compose by running the command:"
echo ""

echo "sh ${AEM_WRAPPER} init"
