#!/usr/bin/env sh

VERSION=${AEMC_VERSION:-"0.8.0"}

# Utilities

# print provisioning step header
step () {
  DATE=$(date "+%Y-%m-%d %H:%M:%S")
  echo "[$DATE]" "$@"
}

# check last command
clc () {
  if [ "$?" -ne 0 ]; then
    exit "$?"
  fi
}

# format seconds to more nice format
duration () {
  T=$1
  D=$((T/60/60/24))
  H=$((T/60/60%24))
  M=$((T/60%60))
  S=$((T%60))
  if [ $D -gt 0 ]; then
      printf '%d day(s) ' $D
  fi
  if [ $H -gt 0 ]; then
      printf '%d hour(s) ' $H
  fi
  if [ $M -gt 0 ]; then
      printf '%d minute(s) ' $M
  fi
  if [ $D -gt 0 ] || [ $H -gt 0 ] || [ $M -gt 0 ]; then
      printf 'and '
  fi
  printf '%d second(s)\n' $S
}

detectOs () {
    case "$(uname)" in
      'Linux')
        echo "linux"
        ;;
      'Darwin')
        echo "darwin"
        ;;
      *)
        echo "windows"
        ;;
    esac
}

detectArch () {
  uname -m
}

downloadFile () {
  URL=$1
  FILE=$2
  if [ ! -f "${FILE}" ]; then
      mkdir -p "$(dirname "$FILE")"
      curl -o "$FILE" -OJL "$URL"
  fi
}

# Download or use installed tool

OS=$(detectOs)
ARCH=$(detectArch)

AEM_DIR="aem"
HOME_DIR="${AEM_DIR}/home"
DOWNLOAD_DIR="${HOME_DIR}/tmp"

BIN_DOWNLOAD_NAME="aemc-cli"
BIN_DOWNLOAD_URL="https://github.com/wttech/aemc/releases/download/v${VERSION}/${BIN_DOWNLOAD_NAME}_${OS}_${ARCH}.tar.gz"
BIN_ROOT="${DOWNLOAD_DIR}/${BIN_DOWNLOAD_NAME}/${VERSION}"
BIN_ARCHIVE_FILE="${BIN_ROOT}/${BIN_DOWNLOAD_NAME}.tar.gz"
BIN_ARCHIVE_DIR="${BIN_ROOT}/${BIN_DOWNLOAD_NAME}"
BIN_NAME="aem"
BIN_EXEC_FILE="${BIN_ARCHIVE_DIR}/${BIN_NAME}"

if [ "${VERSION}" != "installed" ] ; then
  if [ ! -f "${BIN_EXEC_FILE}" ]; then
    mkdir -p "${BIN_ARCHIVE_DIR}"
    downloadFile "${BIN_DOWNLOAD_URL}" "${BIN_ARCHIVE_FILE}"
    tar -xf "${BIN_ARCHIVE_FILE}" -C "${BIN_ARCHIVE_DIR}"
    chmod +x "${BIN_EXEC_FILE}"
  fi
  aem() {
      "./${BIN_EXEC_FILE}" "$@"
  }
fi
