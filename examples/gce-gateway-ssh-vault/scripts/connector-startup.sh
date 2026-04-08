#!/bin/bash
set -e

ACCESS_TOKEN=$(curl -sf -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/attributes/connector-access-token)

REFRESH_TOKEN=$(curl -sf -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/attributes/connector-refresh-token)

TWINGATE_URL=$(curl -sf -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/attributes/connector-url)

curl "https://binaries.twingate.com/connector/setup.sh" | \
  sudo TWINGATE_ACCESS_TOKEN="$ACCESS_TOKEN" \
       TWINGATE_REFRESH_TOKEN="$REFRESH_TOKEN" \
       TWINGATE_URL="$TWINGATE_URL" \
       bash
