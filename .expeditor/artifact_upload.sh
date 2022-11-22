#!/bin/bash

set -eou pipefail

version=$(cat VERSION)

cd cli && make build

  jfrog rt u \
  --apikey="${ARTIFACTORY_TOKEN}" \
  --url=https://artifactory-internal.ps.chef.co/artifactory \
  --target-props "project=license-audit;version=${version};os=linux;arch=amd64" \
  "cli/bin/license-audit" \
  "go-binaries-local/license-audit/${version}/linux/amd64/license-audit"
