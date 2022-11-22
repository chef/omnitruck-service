#!/bin/sh
set -evx

version=$(cat VERSION)

sed -i -r "s/^(\\s*)VERSION: \".+\"/\\1VERSION: \"$version\"/" .expeditor/build.docker.yml
sed -i -r "s/^(.*builder:).*/\\1$version/" builder/terraform/variables.tf
