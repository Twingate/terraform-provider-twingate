#!/bin/bash
set -e
mkdir -p /etc/twingate/

{
  echo TWINGATE_URL="${url}"
  echo TWINGATE_ACCESS_TOKEN="${access_token}"
  echo TWINGATE_REFRESH_TOKEN="${refresh_token}"
} > /etc/twingate/connector.conf

sudo systemctl enable --now twingate-connector