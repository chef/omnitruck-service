#!/bin/sh
set -evx

VERSION=$(cat VERSION)

sed -i -r "s/^(\\s*)VERSION: \".+\"/\\1VERSION: \"$VERSION\"/" .expeditor/build.docker.yml
