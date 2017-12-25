#!/bin/bash

# Exit on error. Append "|| true" if you expect an error.
set -o errexit
# Do not allow use of undefined vars. Use ${VAR:-} to use an undefined VAR
set -o nounset
# Catch the error in pipeline.
set -o pipefail

MAKR_RULES_ROOT=$(dirname "${BASH_SOURCE}")
VERBOSE="${VERBOSE:-1}"
source "${MAKR_RULES_ROOT}/lib/init.sh"

if [[ ! -d "${PRJ_ROOT}/hack/make-rules" ]]; then
    log::error_exit "Please check out make-rules in hack/"
fi

# GNU sed
SED_I=(sed -i)
if [[ $(uname -s) == "Darwin" && -z "$(sed --version 2>&1 | grep "GNU")" ]]; then
    # BSD sed
    SED_I=(sed -i '') 
fi

cd ${MAKR_RULES_ROOT}
MAKEFILE="${PRJ_ROOT}/Makefile"
cp Makefile.tmpl ${MAKEFILE}

[[ -d "${PRJ_ROOT}/cmd" ]] && cmds=($(ls ${PRJ_ROOT}/cmd))
[[ -d "${PRJ_ROOT}/build" ]] && builds=($(ls ${PRJ_ROOT}/build))
prj=$(basename ${PRJ_ROOT})

"${SED_I[@]}" "s|__CMD_TARGETS__|\$(addprefix cmd/,${cmds[*]-})|g" ${MAKEFILE}
"${SED_I[@]}" "s|__DOCKER_TARGETS__|\$(addprefix build/,${builds[*]-})|g" ${MAKEFILE}
"${SED_I[@]}" "s|__PROJECT_PREFIX__|${prj}-|g" ${MAKEFILE}

log::status "Genenrate Makefile successfully"
