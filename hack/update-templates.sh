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

set -o errexit
set -o nounset
set -o pipefail

source "$(go run knative.dev/hack/cmd/script library.sh)"

readonly FUNC_BINARY_DIR="$(mktemp -d ${REPO_ROOT_DIR}/tmpfuncdir.XXXXXX)"

cleanup() {
  rm -rf "${FUNC_BINARY_DIR}"
}

trap "cleanup" EXIT SIGINT

function resolveFuncBinaryName() {
  ARCH=""
  case $(uname -m) in
      i386)   ARCH="386" ;;
      i686)   ARCH="386" ;;
      x86_64) ARCH="amd64" ;;
      arm)    ARCH="amd64" ;;
      arm64)  ARCH="amd64" ;;
      *) echo "** Unknown architecture '$(uname -m)'" ; exit 1 ;;
  esac

  BINARY=""
  case "${OSTYPE}" in
    darwin*) BINARY="func_darwin_${ARCH}" ;;
    linux*)  BINARY="func_linux_${ARCH}" ;;
    *) echo "** Internal error in library.sh, unknown OS '${OSTYPE}'" ; exit 1 ;;
  esac

  echo "${BINARY}"
}

function resolveFuncBinaryUrl() {
  # e.g.
  # https://github.com/knative/func/releases/download/knative-v1.15.0/func_linux_amd64
  VERSION=$(cat "${REPO_ROOT_DIR}/hack/func-version.txt")
  echo "https://github.com/knative/func/releases/download/${VERSION}/$(resolveFuncBinaryName)"
}

# download the binary into the temporary directory
echo "Downloading func binary from $(resolveFuncBinaryUrl)"
curl -sL "$(resolveFuncBinaryUrl)" -o "${FUNC_BINARY_DIR}/func"
chmod +x "${FUNC_BINARY_DIR}/func"
# add the func binary to the PATH
export PATH="${FUNC_BINARY_DIR}:${PATH}"

TEMPLATES_DIR=${REPO_ROOT_DIR}/backstage/templates
SKELETONS_DIR=${TEMPLATES_DIR}/skeletons

# iterate over the list of lang+template entries (skip the first line, which is the header)
TEMPLATES_LINES=$(func templates | tail -n +2)
IFS=$'\n' read -d '' -r -a TEMPLATE_TUPLES <<< "$TEMPLATES_LINES" || true
for tuple in "${TEMPLATE_TUPLES[@]}"
do
  IFS=' ' read -r -a tuple <<< "$tuple"
  LANG=${tuple[0]}
  TEMPLATE=${tuple[1]}
  NAME="$LANG-$TEMPLATE"

  # remove existing skeleton
  rm -rf "${SKELETONS_DIR}/${NAME}" || true

  # create skeleton
  echo "Creating function for language: $LANG, template: $TEMPLATE"
  func create -l $LANG -t $TEMPLATE "${SKELETONS_DIR}/${NAME}"

  # replace the line in func.yaml that starts with "name:" with "name: $NAME"
  sed -i "s/^name: .*/name: ${NAME}/" "${SKELETONS_DIR}/${NAME}/func.yaml"

  # replace the line in func.yaml that starts with "created" with "created: 2024-01-01T00:00:00.000000+00:00"
  sed -i "s/^created: .*/created: 2024-01-01T00:00:00.000000+00:00/" "${SKELETONS_DIR}/${NAME}/func.yaml"

  # remove the .func directory
  rm -rf "${SKELETONS_DIR}/${NAME}/.func" || true

  # remove and recreate the template yaml file
  TEMPLATE_FILE="${TEMPLATES_DIR}/${NAME}.yaml"
  rm $TEMPLATE_FILE || true
  echo "Generating template yaml file for language: $LANG, template: $TEMPLATE at $TEMPLATE_FILE"
  cat "${REPO_ROOT_DIR}/hack/backstage-template-template.yaml" | LANG=$LANG TEMPLATE=$TEMPLATE envsubst > $TEMPLATE_FILE
done

# generate location.yaml file
LOCATION_FILE="${TEMPLATES_DIR}/location.yaml"
rm $LOCATION_FILE || true
cp "${REPO_ROOT_DIR}/hack/backstage-location-template.yaml" $LOCATION_FILE
for tuple in "${TEMPLATE_TUPLES[@]}"
do
  IFS=' ' read -r -a tuple <<< "$tuple"
  LANG=${tuple[0]}
  TEMPLATE=${tuple[1]}
  NAME="$LANG-$TEMPLATE"
  echo "  - ./${NAME}.yaml" >> $LOCATION_FILE
done

echo "Done"
