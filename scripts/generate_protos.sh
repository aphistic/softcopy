#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

PROTO_PATHS=(
    "pkg/proto"
)

for PROTO_PATH in "${PROTO_PATHS[@]}"
do
    cd ${SCRIPT_DIR}/../${PROTO_PATH}
    echo "Generating ${PROTO_PATH}"
    ./generate.sh
    echo ""
done
