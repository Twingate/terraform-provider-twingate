#!/bin/bash
set -eo pipefail

SETUP_SCRIPT=$(mktemp)

# Retry because outbound access through the NAT gateway can lag behind boot.
# Download to a file so a retry restarts the file instead of appending to it,
# and nothing runs until the whole script has arrived.
curl -sSfL --retry 10 --retry-delay 5 --retry-all-errors --connect-timeout 10 \
  -o "$SETUP_SCRIPT" "https://binaries.twingate.com/connector/setup.sh"

sudo TWINGATE_ACCESS_TOKEN="${access_token}" \
     TWINGATE_REFRESH_TOKEN="${refresh_token}" \
     TWINGATE_URL="${twingate_url}" \
     bash "$SETUP_SCRIPT"

rm -f "$SETUP_SCRIPT"
