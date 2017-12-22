#!/bin/bash

# Exit on error. Append "|| true" if you expect an error.
set -o errexit
# Do not allow use of undefined vars. Use ${VAR:-} to use an undefined VAR
set -o nounset
# Catch the error in pipeline.
set -o pipefail

MAKR_RULES_ROOT=$(dirname "${BASH_SOURCE}")/..
VERBOSE="${VERBOSE:-1}"
source "${MAKR_RULES_ROOT}/lib/init.sh"

rm -rf ${PRJ_OUTPUT_BINPATH}
