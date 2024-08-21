#!/usr/bin/env bash

# Copyright 2024 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

export GO111MODULE=on

source "$(dirname "${BASH_SOURCE[0]}")/../vendor/knative.dev/hack/e2e-tests.sh"

function test_plugins() {
  echo "Building and testing Backstage plugins"
  pushd ./backstage
  yarn --prefer-offline --frozen-lockfile
  npm install @backstage/cli -g
  yarn backstage-cli repo lint
  yarn tsc
  yarn test --watchAll=false
  yarn build:all
  popd
}

# Script entry point.

# TODO: to be enabled when we need to provision a cluster
## initialize "$@"

test_plugins || fail_test
