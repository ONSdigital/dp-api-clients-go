#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-api-clients-go
  make build
popd