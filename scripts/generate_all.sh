#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

${SCRIPT_DIR}/generate_vaults.sh
${SCRIPT_DIR}/generate_protos.sh
${SCRIPT_DIR}/generate_mocks.sh
