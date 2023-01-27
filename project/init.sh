#!/usr/bin/env sh

VERSION=${AEMC_VERSION:-"0.12.3"}
SOURCE_URL="https://raw.githubusercontent.com/wttech/aemc/v${VERSION}/project"

AEM_WRAPPER="aemw"

TARGET_AEM_DIR="instance" # change it if you don't want AEMC to be installed in 'aem' folder

TARGET_SCRIPT_DIR="${TARGET_AEM_DIR}/script"
TARGET_HOME_DIR="${TARGET_AEM_DIR}/home"
TARGET_DEFAULT_DIR="${TARGET_AEM_DIR}/default"
TARGET_DEFAULT_CONFIG_DIR="${TARGET_DEFAULT_DIR}/etc"
TARGET_LIB_DIR="${TARGET_HOME_DIR}/lib"

DOWNLOAD_AEM_DIR="aem"
SCRIPT_DIR="${DOWNLOAD_AEM_DIR}/script"
HOME_DIR="${DOWNLOAD_AEM_DIR}/home"
DEFAULT_DIR="${DOWNLOAD_AEM_DIR}/default"
DEFAULT_CONFIG_DIR="${DEFAULT_DIR}/etc"
LIB_DIR="${HOME_DIR}/lib"

if [ -f "$AEM_WRAPPER" ]; then
 echo "The project contains already AEM Compose!"
 exit 1
fi

echo "Downloading AEM Compose Files"
echo ""

mkdir -p "${TARGET_SCRIPT_DIR}" "${TARGET_HOME_DIR}" "${TARGET_DEFAULT_CONFIG_DIR}" "${TARGET_LIB_DIR}"

curl -s "${SOURCE_URL}/${DEFAULT_CONFIG_DIR}/aem.yml" -o "${TARGET_DEFAULT_CONFIG_DIR}/aem.yml"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/deploy.sh" -o "${TARGET_SCRIPT_DIR}/deploy.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/destroy.sh" -o "${TARGET_SCRIPT_DIR}/destroy.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/down.sh" -o "${TARGET_SCRIPT_DIR}/down.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/resetup.sh" -o "${TARGET_SCRIPT_DIR}/resetup.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/restart.sh" -o "${TARGET_SCRIPT_DIR}/restart.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/setup.sh" -o "${TARGET_SCRIPT_DIR}/setup.sh"
curl -s "${SOURCE_URL}/${SCRIPT_DIR}/up.sh" -o "${TARGET_SCRIPT_DIR}/up.sh"
curl -s "${SOURCE_URL}/${DOWNLOAD_AEM_DIR}/api.sh" -o "${TARGET_AEM_DIR}/api.sh"
curl -s "${SOURCE_URL}/${AEM_WRAPPER}" -o "${AEM_WRAPPER}"

echo "Downloading & Testing AEM Compose CLI"
echo ""
# <https://stackoverflow.com/a/57766728>
if [ "$(uname)" = "Darwin" ]; then
  sed -i '' 's/AEM_DIR="aem"/AEM_DIR="'"$TARGET_AEM_DIR"'"/g' "$AEM_WRAPPER"
  sed -i '' 's/AEM_DIR="aem"/AEM_DIR="'"$TARGET_AEM_DIR"'"/g' "$TARGET_AEM_DIR/api.sh"
else
  sed -i 's/AEM_DIR="aem"/AEM_DIR="'"$TARGET_AEM_DIR"'"/g' "$AEM_WRAPPER"
  sed -i 's/AEM_DIR="aem"/AEM_DIR="'"$TARGET_AEM_DIR"'"/g' "$TARGET_AEM_DIR/api.sh"
fi

chmod +x "${AEM_WRAPPER}"
sh ${AEM_WRAPPER} version

echo "Success! Now initialize AEM Compose by running the command:"
echo ""

echo "sh ${AEM_WRAPPER} init"
