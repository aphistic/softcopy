#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

VAULT_PATHS=(
    "internal/app/softcopy-server/uiserver/backend"
    "internal/app/softcopy-server/uiserver/frontend"
    "internal/pkg/storage/data/sqlite/migrations"
)

for VAULT_PATH in "${VAULT_PATHS[@]}"
do
    cd ${SCRIPT_DIR}/../${VAULT_PATH}
    echo "Generating ${VAULT_PATH}"
    go generate
    echo ""
done
