#!/bin/bash

# Exit on error. Append "|| true" if you expect an error.
set -o errexit
# Do not allow use of undefined vars. Use ${VAR:-} to use an undefined VAR
set -o nounset
# Catch the error in pipeline.
set -o pipefail

# The root of the build/dist directory
# make-rules should be palced in project/hack/
# so the project root is project/hack/make-rules/lib/../../../
MAKE_RULES_ROOT="$(cd "$(dirname "${BASH_SOURCE}")/.." && pwd -P)"
PRJ_ROOT="$(cd "$(dirname "${BASH_SOURCE}")/../../.." && pwd -P)"
PRJ_CMDPATH="${PRJ_ROOT}/cmd"
PRJ_OUTPUT_BINPATH="${PRJ_ROOT}/bin"

GO_ONBUILD_IMAGE="${GO_ONBUILD_IMAGE:-golang:1.9.2-alpine3.6}"
COLOR_LOG=true

source "${MAKE_RULES_ROOT}/lib/util.sh"
source "${MAKE_RULES_ROOT}/lib/logging.sh"

log::install_errexit

source "${MAKE_RULES_ROOT}/lib/version.sh"
source "${MAKE_RULES_ROOT}/lib/golang.sh"
source "${MAKE_RULES_ROOT}/lib/docker.sh"

PRJ_OUTPUT_HOSTBIN="${PRJ_OUTPUT_BINPATH}/$(util::host_platform)"
