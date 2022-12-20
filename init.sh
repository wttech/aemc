#!/usr/bin/env sh

VERSION=${AEMC_VERSION:-"0.5.12"}
SOURCE_URL=https://raw.githubusercontent.com/wttech/aemc/v${VERSION}

AEM_WRAPPER=aemw
AEM_DIR=aem
SCRIPT_DIR=${AEM_DIR}/script
HOME_DIR=${AEM_DIR}/home
CONFIG_FILE=${HOME_DIR}/aem.yml

if [ -f "$AEM_WRAPPER" ]; then
  echo "The project contains already AEM Compose!"
  exit 1
fi

echo "Downloading AEM Compose Files"
echo ""

mkdir -p "$SCRIPT_DIR" "$HOME_DIR"
curl -s ${SOURCE_URL}/${SCRIPT_DIR}/destroy.sh -o ${SCRIPT_DIR}/destroy.sh
curl -s ${SOURCE_URL}/${SCRIPT_DIR}/down.sh -o ${SCRIPT_DIR}/down.sh
curl -s ${SOURCE_URL}/${SCRIPT_DIR}/resetup.sh -o ${SCRIPT_DIR}/resetup.sh
curl -s ${SOURCE_URL}/${SCRIPT_DIR}/restart.sh -o ${SCRIPT_DIR}/restart.sh
curl -s ${SOURCE_URL}/${SCRIPT_DIR}/setup.sh -o ${SCRIPT_DIR}/setup.sh
curl -s ${SOURCE_URL}/${SCRIPT_DIR}/up.sh -o ${SCRIPT_DIR}/up.sh
curl -s ${SOURCE_URL}/${AEM_DIR}/api.sh -o ${AEM_DIR}/api.sh
curl -s ${SOURCE_URL}/${AEM_WRAPPER} -o ${AEM_WRAPPER}

echo "Downloading & Running AEM Compose CLI"
echo ""

chmod +x "${AEM_WRAPPER}"
sh ${AEM_WRAPPER} version

echo "Scaffolding AEM Compose configuration file"
echo ""

./${AEM_WRAPPER} config init

echo "Initialized AEM Compose"
echo ""

echo "The next step is instructing the tool where AEM files are located (JAR or SDK ZIP, license)."
echo "Adjust the configuration file '${CONFIG_FILE}' by updating properties: 'dist_path', 'license_path'."
echo "Finally, use control scripts to manage AEM instances:"
echo ""

echo "sh aemw [setup|resetup|up|down|restart]"

echo ""
echo "It is also possible to run individual AEM Compose CLI commands separately."
echo "Discover available commands by running:"
echo ""

echo "sh aemw --help"
