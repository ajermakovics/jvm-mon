#!/bin/bash
# Build and run from source. Builds the agent jar if missing.

set -euo pipefail
cd "$(dirname "$0")"

if [ ! -f build/libs/jvm-mon-go.jar ]; then
  ./make-agent.sh
fi

go build -o build/jvm-mon-go && echo Built && ./build/jvm-mon-go "$@"
