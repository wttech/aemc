#!/usr/bin/env sh

AEM_WRAPPER="aemw"

if [ -f "$AEM_WRAPPER" ]; then
  echo ""
  echo "The project already contains AEM Compose!"
  exit 0
fi

SOURCE_URL="https://raw.githubusercontent.com/wttech/aemc/main/pkg/project/common"
curl -s "${SOURCE_URL}/${AEM_WRAPPER}" -o "${AEM_WRAPPER}"

echo ""
echo "Downloading & Testing AEM Compose CLI"
echo ""

chmod +x "${AEM_WRAPPER}"
sh ${AEM_WRAPPER} version

echo ""
echo "Success! Now initialize the project files by running command below:"
echo ""

echo "sh ${AEM_WRAPPER} init"
