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

readonly SKIP_INITIALIZE=${SKIP_INITIALIZE:-false}
readonly LOCAL_DEVELOPMENT=${LOCAL_DEVELOPMENT:-false}
export KO_FLAGS="${KO_FLAGS:-}"

repo_root_dir=$(dirname "$(realpath "${BASH_SOURCE[0]}")")/..

source "${repo_root_dir}"/vendor/knative.dev/hack/e2e-tests.sh
source "${repo_root_dir}"/hack/control-plane.sh
source "${repo_root_dir}"/hack/artifacts-env.sh

# If gcloud is not available make it a no-op, not an error.
which gcloud &>/dev/null || gcloud() { echo "[ignore-gcloud $*]" 1>&2; }

# Use GNU tools on macOS. Requires the 'grep' and 'gnu-sed' Homebrew formulae.
if [ "$(uname)" == "Darwin" ]; then
  sed=gsed
  grep=ggrep
fi

# Latest release. If user does not supply this as a flag, the latest tagged release on the current branch will be used.
readonly LATEST_RELEASE_VERSION=$(latest_version)
readonly PREVIOUS_RELEASE_URL="${PREVIOUS_RELEASE_URL:-"https://github.com/knative-extensions/backstage-plugins/releases/download/${LATEST_RELEASE_VERSION}"}"

readonly EVENTING_CONFIG=${EVENTING_CONFIG:-"./third_party/eventing-latest/"}

# Vendored eventing test images.
readonly VENDOR_EVENTING_TEST_IMAGES="vendor/knative.dev/eventing/test/test_images/"

export SYSTEM_NAMESPACE="knative-eventing"
export CLUSTER_SUFFIX=${CLUSTER_SUFFIX:-"cluster.local"}

function knative_setup() {
  knative_eventing
  return $?
}

function knative_teardown() {
  if ! is_release_branch; then
    echo ">> Delete Knative Eventing from HEAD"
    pushd .
    kubectl delete --ignore-not-found -f "${EVENTING_CONFIG}"
    popd || fail_test "Failed to set up Eventing"
  else
    echo ">> Delete Knative Eventing from ${KNATIVE_EVENTING_RELEASE}"
    kubectl delete --ignore-not-found -f "${KNATIVE_EVENTING_RELEASE}"
  fi
}

function knative_eventing() {
  if ! is_release_branch; then
    echo ">> Install Knative Eventing from latest - ${EVENTING_CONFIG}"
    kubectl apply -f "${EVENTING_CONFIG}/eventing-crds.yaml"
    kubectl apply -f "${EVENTING_CONFIG}/eventing-core.yaml"
  else
    echo ">> Install Knative Eventing from ${KNATIVE_EVENTING_RELEASE}"
    kubectl apply -f "${KNATIVE_EVENTING_RELEASE}"
  fi

  ! kubectl patch horizontalpodautoscalers.autoscaling -n knative-eventing eventing-webhook -p '{"spec": {"minReplicas": '${REPLICAS}'}}'

  # Publish test images.
  echo ">> Publishing test images from eventing"
  ./test/upload-test-images.sh ${VENDOR_EVENTING_TEST_IMAGES} e2e || fail_test "Error uploading test images"
}

function build_components_from_source() {
  header "Building components from source"

  [ -f "${BACKEND_ARTIFACT}" ] && rm "${BACKEND_ARTIFACT}"
  # TODO: later
  #[ -f "${BACKEND_POST_INSTALL_ARTIFACT}" ] && rm "${BACKEND_POST_INSTALL_ARTIFACT}"

  header "Backend setup"
  backend_setup || fail_test "Failed to set up backend"

  return $?
}

function install_latest_release() {
  echo "Installing latest release from ${PREVIOUS_RELEASE_URL}"

  ko apply ${KO_FLAGS} -f ./test/config/ || fail_test "Failed to apply test configurations"

  kubectl apply -f "${PREVIOUS_RELEASE_URL}/${BACKEND_ARTIFACT}" || return $?

  # Restore test config.
  kubectl replace -f ./test/config/100-config-tracing.yaml
}

function install_head() {
  echo "Installing head"

  kubectl apply -f "${BACKEND_ARTIFACT}" || return $?
  # TODO: later
  # kubectl apply -f "${BACKEND_POST_INSTALL_ARTIFACT}" || return $?

  # Restore test config.
  kubectl replace -f ./test/config/100-config-tracing.yaml
}

function test_setup() {
  header "Test setup"

  build_components_from_source || return $?

  install_head || return $?

  wait_until_pods_running knative-eventing || fail_test "System did not come up"

  # Apply test configurations
  ko apply ${KO_FLAGS} -f ./test/config/ || fail_test "Failed to apply test configurations"
}

function test_teardown() {
  kubectl delete --ignore-not-found -f "${BACKEND_ARTIFACT}" || fail_test "Failed to tear down backend"
}

function export_logs_continuously() {

  labels=("eventmesh-backend")

  mkdir -p "$ARTIFACTS/${SYSTEM_NAMESPACE}"

  for deployment in "${labels[@]}"; do
    kubectl logs -n ${SYSTEM_NAMESPACE} -f -l=app="$deployment" >"$ARTIFACTS/${SYSTEM_NAMESPACE}/$deployment" 2>&1 &
  done
}

function save_release_artifacts() {
  # Copy our release artifacts into artifacts, so that release artifacts of a PR can be tested and reviewed without
  # building the project from source.
  cp "${BACKEND_ARTIFACT}" "${ARTIFACTS}/${BACKEND_ARTIFACT}" || return $?
}
