#!/usr/bin/env bash
: ${WFSM_TAG:="0.0.1-dev.23"}
: ${WFSM_ARCH:=$(arch)}
: ${WFSM_OS:=$(echo $(uname -s) | tr '[:upper:]' '[:lower:]')}
: ${WFSM_TARGET:=$(pwd)/bin}

WFSM_ARCHIVE_URL="https://github.com/agntcy/workflow-srv-mgr/releases/download/v${WFSM_TAG}/wfsm${WFSM_TAG}_${WFSM_OS}_${WFSM_ARCH}.tar.gz"
echo "Installing the Workflow Server Manager tool:"
echo "OS:" "$WFSM_OS"
echo "ARCH:" "$WFSM_ARCH"
echo "AG:" "$WFSM_TAG"
echo "ARCHIVE_URL:" "$WFSM_ARCHIVE_URL"

set -e
rm -f "$WFSM_TARGET/wfsm"

# Create the target directory if it doesn't exist
mkdir -p "$WFSM_TARGET"

# Check if the version exists
if ! curl --head --fail --output /dev/null "$WFSM_ARCHIVE_URL" 2> /dev/null;
 then
  echo "Version not found"
  exit 1
fi

# Download and extract the archive
curl -s -S -L "$WFSM_ARCHIVE_URL" | tar -xf - -C "$WFSM_TARGET"

# Make the binary executable
chmod +x "$WFSM_TARGET/wfsm"

Echo "Installation complete. The 'wfsm' binary is located at $WFSM_TARGET/wfsm"
