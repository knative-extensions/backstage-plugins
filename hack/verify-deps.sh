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

set -o errexit
set -o nounset
set -o pipefail

source "$(go run knative.dev/hack/cmd/script library.sh)"

readonly TMP_DIFFROOT="$(mktemp -d ${REPO_ROOT_DIR}/tmpdepsdiffroot.XXXXXX)"

cleanup() {
  rm -rf "${TMP_DIFFROOT}"
}

trap "cleanup" EXIT SIGINT

# Save working tree state
cp -aR "${REPO_ROOT_DIR}/go.sum" "${TMP_DIFFROOT}"

ret=0
echo "Checking generated FS"

"${REPO_ROOT_DIR}/hack/update-deps.sh"
echo "Diffing ${REPO_ROOT_DIR} against freshly generated codegen"
diff -Nupr --no-dereference "${REPO_ROOT_DIR}/go.sum" "${TMP_DIFFROOT}/go.sum" || ret=1

# Restore working tree state
rm -fr "${REPO_ROOT_DIR}/go.sum"
cp -aR "${TMP_DIFFROOT}"/* "${REPO_ROOT_DIR}"

if [[ $ret -eq 0 ]]
then
  echo "${REPO_ROOT_DIR} up to date."
else
  echo "ERROR: ${REPO_ROOT_DIR} is out of date. Please run ./hack/update-deps.sh"
  exit 1
fi

