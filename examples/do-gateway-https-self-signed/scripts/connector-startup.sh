#!/bin/bash
set -e

curl "https://binaries.twingate.com/connector/setup.sh" | \
  sudo TWINGATE_ACCESS_TOKEN="${access_token}" \
       TWINGATE_REFRESH_TOKEN="${refresh_token}" \
       TWINGATE_URL="${twingate_url}" \
       bash
