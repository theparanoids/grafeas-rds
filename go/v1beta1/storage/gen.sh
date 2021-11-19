#!/usr/bin/env bash
#
# The copyright header is deliberately placed below the shebang line because
# the OS only looks at the first two bytes to determine which executable to use.
#
# Copyright Yahoo 2021
# Licensed under the terms of the Apache License 2.0.
# See LICENSE file in project root for terms.
#
# This script should only be invoked by "go generate"
# because it makes sure that the current working directory is the project root,
# so the tool binaries will be placed inside the bin directory of this project,
# and they will be git-ignored.
set -euo pipefail

main() {
  # Since different projects may have different versions of mockgen,
  # we first install mockgen of the version specified in the go.mod of this project,
  # and the binary will be placed under the bin directory of this project and git-ignored.
  GOBIN=$(pwd)/bin
  export GOBIN
  go install github.com/golang/mock/mockgen

  # Make sure that we are using the mockgen which is just built to generate mocks code.
  export PATH=$GOBIN:$PATH
  mockgen -package mocks -destination ../mocks/driver_mock.go database/sql/driver Driver
  mockgen -package mocks -destination ../mocks/conn_pool_mgr_mock.go . ConnPoolMgr
  mockgen -package mocks -destination ../mocks/storage_mock.go . Storage
  mockgen -package mocks -destination ../mocks/credentials_creator_mock.go . CredentialsCreator
  mockgen -package storage -destination storage_creator_mock_test.go . StorageCreator
}

main
