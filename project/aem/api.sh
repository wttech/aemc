#!/usr/bin/env sh

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
