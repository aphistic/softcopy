#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

MOCK_PATHS=(
    "pkg/proto/mock"
)

for MOCK_PATH in "${MOCK_PATHS[@]}"
do
    cd ${SCRIPT_DIR}/../${MOCK_PATH}
    echo "Generating ${MOCK_PATH}"
    go generate
    echo ""
done
