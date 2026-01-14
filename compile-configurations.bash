#!/bin/bash

set -euo pipefail

echo "Warning: this script does NOT check that all environment \
variables are defined and will substitute variables with empty strings."

find . -name '*.template.conf' -type f -print0 | while IFS= read -r -d '' filepath; do
    compiled_filename=$(echo $filepath | sed 's/\.template/\.compiled/')
    cat $filepath | envsubst > $compiled_filename

    echo "    Compiled '$filepath' to '$compiled_filename'"
done