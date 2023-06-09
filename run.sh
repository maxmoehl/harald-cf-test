#!/usr/bin/env ash

set -exo pipefail

admin_port="$(yq '.admin.address.socket_address.port_value' "/etc/cf-assets/envoy_config/envoy.yaml")"

# envoy takes a moment to come up
curl -ifX POST \
  --retry 10 --retry-all-errors --retry-delay 1 \
  "127.0.0.1:${admin_port}/drain_listeners"

envoy-to-harald "/etc/cf-assets/envoy_config/envoy.yaml" > "/etc/cf-assets/harald.yml"

harald "/etc/cf-assets/harald.yml" &

test-app
