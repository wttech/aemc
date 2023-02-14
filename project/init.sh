#!/usr/bin/env sh

AEM_WRAPPER="aemw"

if [ -f "$AEM_WRAPPER" ]; then
  echo "The project contains already AEM Compose!"
  exit 1
fi

SOURCE_URL="https://raw.githubusercontent.com/wttech/aemc/main/project"
curl -s "${SOURCE_URL}/${AEM_WRAPPER}" -o "${AEM_WRAPPER}"

echo "Downloading & Testing AEM Compose CLI"
echo ""

chmod +x "${AEM_WRAPPER}"
sh ${AEM_WRAPPER} version

echo "Success! Now initialize AEM Compose by running the command:"
echo ""

echo "sh ${AEM_WRAPPER} init --project-type [classic|cloud]"
