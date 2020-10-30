#!/bin/bash

MODULE="github.com/aphistic/softcopy"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

MOCK_PATHS=(
    "pkg/proto"
)

for MOCK_PATH in "${MOCK_PATHS[@]}"
do
    echo "Generating ${MOCK_PATH}"
    docker run --rm -i -v ${SCRIPT_DIR}/..:/app efritz/go-mockgen \
        ${MODULE}/${MOCK_PATH} \
        -d ${MOCK_PATH}/mock \
        -f \
        -p protomock \
        -i SoftcopyClient -i SoftcopyAdminClient
    echo ""
done
