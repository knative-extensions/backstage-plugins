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

source "$(dirname "$0")/e2e-common.sh"

# Script entry point.
initialize "$@"

echo "Running E2E tests"
go_test_e2e -timeout=1h -parallel=20 ./backends/tests/e2e/... || fail_test
success
