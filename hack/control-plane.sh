#!/usr/bin/env bash

# Copyright 2020 The Knative Authors
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

source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")"/label.sh

readonly BACKEND_CONFIG_DIR=backends/config/100-eventmesh

# Note: do not change this function name, it's used during releases.
function backend_setup() {
  ko resolve ${KO_FLAGS} -Rf "${BACKEND_CONFIG_DIR}" | "${LABEL_YAML_CMD[@]}" >>"${BACKEND_ARTIFACT}" || return $?
}
