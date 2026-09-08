#!/bin/bash
set -e

# Wait until the machine can reach the Twingate binaries server. This is a work around for making sure outbound connectivity is available
until curl -sf --connect-timeout 5 https://binaries.twingate.com > /dev/null 2>&1; do
  sleep 2
done

until sudo apt-get update -y; do sleep 5; done

curl "https://binaries.twingate.com/connector/setup.sh" | \
  sudo TWINGATE_ACCESS_TOKEN="${access_token}" \
       TWINGATE_REFRESH_TOKEN="${refresh_token}" \
       TWINGATE_URL="${twingate_url}" \
       bash
