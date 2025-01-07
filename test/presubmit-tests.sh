#!/usr/bin/env bash

# Copyright 2023 The Knative Authors
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

# This script runs the presubmit tests; it is started by prow for each PR.
# For convenience, it can also be executed manually.
# Running the script without parameters, or with the --all-tests
# flag, causes all tests to be executed, in the right order.
# Use the flags --build-tests, --unit-tests and --integration-tests
# to run a specific set of tests.

export GO111MODULE=on
export DISABLE_MD_LINTING=1

source "$(go run knative.dev/hack/cmd/script presubmit-tests.sh)"

function unit_tests() {
  header "Running Go unit tests"
  default_unit_test_runner || fail_test "Go unit tests failed"

  header "Building and testing Backstage plugins"
  pushd ./backstage
  yarn --prefer-offline --frozen-lockfile || fail_test "failed to build plugins"
  npm install @backstage/cli -g
  yarn backstage-cli repo lint || fail_test "failed to initialize repo"
  yarn tsc || fail_test "failed to bundle plugins"
  yarn test --watchAll=false || fail_test "failed to test plugins"
  yarn build:all || fail_test "failed to build all plugins"
  popd
}

main $@
