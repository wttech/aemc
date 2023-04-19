#!/usr/bin/env sh

AEM_WRAPPER="aemw"
TASK_WRAPPER="taskw"

if [ -f "$AEM_WRAPPER" ]; then
  echo "The project contains already AEM Compose!"
  exit 1
fi

SOURCE_URL="https://raw.githubusercontent.com/wttech/aemc/main/pkg/project/common"
curl -s "${SOURCE_URL}/${AEM_WRAPPER}" -o "${AEM_WRAPPER}"
curl -s "${SOURCE_URL}/${TASK_WRAPPER}" -o "${TASK_WRAPPER}"

echo "Downloading & Testing AEM Compose CLI"
echo ""

chmod +x "${AEM_WRAPPER}" "${TASK_WRAPPER}"
sh ${AEM_WRAPPER} version

echo ""
echo "Success! Now initialize the project by running command below:"
echo ""

echo "sh ${AEM_WRAPPER} init"
