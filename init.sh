#!/usr/bin/env sh

SOURCE_URL=https://raw.githubusercontent.com/wttech/aemc/main
AEM_DIR=aem
SCRIPT_DIR=${AEM_DIR}/script
WRAPPER_SCRIPT=aemw

if [ ! -f "$WRAPPER_SCRIPT" ]; then
  echo "Downloading AEM Compose Initial Files"

  mkdir -p "$SCRIPT_DIR" && \
  curl -s ${SOURCE_URL}/${SCRIPT_DIR}/destroy.sh -o ${SCRIPT_DIR}/destroy.sh && \
  curl -s ${SOURCE_URL}/${SCRIPT_DIR}/down.sh -o ${SCRIPT_DIR}/down.sh && \
  curl -s ${SOURCE_URL}/${SCRIPT_DIR}/resetup.sh -o ${SCRIPT_DIR}/resetup.sh && \
  curl -s ${SOURCE_URL}/${SCRIPT_DIR}/restart.sh -o ${SCRIPT_DIR}/restart.sh && \
  curl -s ${SOURCE_URL}/${SCRIPT_DIR}/setup.sh -o ${SCRIPT_DIR}/setup.sh && \
  curl -s ${SOURCE_URL}/${SCRIPT_DIR}/up.sh -o ${SCRIPT_DIR}/up.sh && \
  curl -s ${SOURCE_URL}/${AEM_DIR}/api.sh -o ${AEM_DIR}/api.sh && \
  curl -s ${SOURCE_URL}/${WRAPPER_SCRIPT} -o ${WRAPPER_SCRIPT} && \
fi

echo "Downloading AEM Compose CLI Executable"

chmod +x "${WRAPPER_SCRIPT}"
./${WRAPPER_SCRIPT} config init
./${WRAPPER_SCRIPT} --help
