#!/usr/bin/env bash
# Copyright 2018 The Kubernetes Authors.
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

source common.sh

export TRACE=1
export GO111MODULE=on

fetch_tools
install_kind
build_kb

setup_envs

source "$(pwd)/scripts/setup.sh" ${KIND_K8S_VERSION}
docker pull gcr.io/kubebuilder/kube-rbac-proxy:v0.5.0
kind load docker-image gcr.io/kubebuilder/kube-rbac-proxy:v0.5.0

# remove running containers on exit
function cleanup() {
    kind delete cluster
}

trap cleanup EXIT
go test ./test/e2e/ -v -ginkgo.v
