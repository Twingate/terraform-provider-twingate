#!/bin/bash
set -eo pipefail

# Retry because outbound access through the NAT gateway can lag behind boot.
curl -sSfL --retry 10 --retry-delay 5 --retry-all-errors --connect-timeout 10 \
  "https://binaries.twingate.com/connector/setup.sh" | \
  sudo TWINGATE_ACCESS_TOKEN="${access_token}" \
       TWINGATE_REFRESH_TOKEN="${refresh_token}" \
       TWINGATE_URL="${twingate_url}" \
       bash
