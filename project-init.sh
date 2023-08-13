#!/usr/bin/env sh

AEM_WRAPPER="aemw"
TASK_WRAPPER="taskw"

if [ -f "$AEM_WRAPPER" ]; then
  echo ""
  echo "The project already contains AEM Compose!"
  exit 1
fi

SOURCE_URL="https://raw.githubusercontent.com/wttech/aemc/main/pkg/project/common"
curl -s "${SOURCE_URL}/${AEM_WRAPPER}" -o "${AEM_WRAPPER}"
curl -s "${SOURCE_URL}/${TASK_WRAPPER}" -o "${TASK_WRAPPER}"

echo ""
echo "Downloading & Testing AEM Compose CLI (https://github.com/wttech/aemc)"
echo ""

chmod +x "${AEM_WRAPPER}"
sh ${AEM_WRAPPER} version

echo ""
echo "Downloading & Testing Task Tool (https://github.com/go-task/task)"
echo ""

chmod +x "${TASK_WRAPPER}"
sh ${TASK_WRAPPER} --version

echo ""
echo "Success! Now initialize the project by running command below:"
echo ""

echo "sh ${TASK_WRAPPER} init"
