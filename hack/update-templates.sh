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

ROOT_DIR=$(dirname "$0")/..
TEMPLATES_DIR=${ROOT_DIR}/backstage/templates
SKELETONS_DIR=${TEMPLATES_DIR}/skeletons

# get the list of templates (skip the first line, which is the header)
TEMPLATES_LINES=$(func templates | tail -n +2)
IFS=$'\n' read -d '' -r -a TEMPLATE_TUPLES <<< "$TEMPLATES_LINES" || true
for tuple in "${TEMPLATE_TUPLES[@]}"
do
  IFS=' ' read -r -a tuple <<< "$tuple"
  LANG=${tuple[0]}
  TEMPLATE=${tuple[1]}
  NAME="$LANG-$TEMPLATE"
  rm -rf "${SKELETONS_DIR}/${NAME}" || true
  echo "Creating function for language: $LANG, template: $TEMPLATE"
  func create -l $LANG -t $TEMPLATE "${SKELETONS_DIR}/${NAME}"

  # replace the line in func.yaml that starts with "name:" with "name: $NAME"
  sed -i "s/^name: .*/name: ${NAME}/" "${SKELETONS_DIR}/${NAME}/func.yaml"

  # replace the line in func.yaml that starts with "created" with "created: 2024-01-01T00:00:00.000000+00:00"
  sed -i "s/^created: .*/created: 2024-01-01T00:00:00.000000+00:00/" "${SKELETONS_DIR}/${NAME}/func.yaml"
done

# generate template yaml files
for tuple in "${TEMPLATE_TUPLES[@]}"
do
  IFS=' ' read -r -a tuple <<< "$tuple"
  LANG=${tuple[0]}
  TEMPLATE=${tuple[1]}
  NAME="$LANG-$TEMPLATE"
  OUTFILE="${TEMPLATES_DIR}/${NAME}.yaml"
  rm $OUTFILE || true
  echo "Generating template yaml file for language: $LANG, template: $TEMPLATE at $OUTFILE"
  export LANG
  export TEMPLATE
  cat "${ROOT_DIR}/hack/backstage-template-template.yaml" | envsubst > $OUTFILE
done

# generate location.yaml file
OUTFILE="${TEMPLATES_DIR}/location.yaml"
rm $OUTFILE || true
cp "${ROOT_DIR}/hack/backstage-location-template.yaml" $OUTFILE
for tuple in "${TEMPLATE_TUPLES[@]}"
do
  IFS=' ' read -r -a tuple <<< "$tuple"
  LANG=${tuple[0]}
  TEMPLATE=${tuple[1]}
  NAME="$LANG-$TEMPLATE"
  echo "  - ./${NAME}.yaml" >> $OUTFILE
done

echo "Done"
