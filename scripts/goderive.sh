#!/bin/bash
set -e
ARG=$1
if [ -z "$ARG" ]; then
    ARG="."
fi

go install github.com/ng-vu/sqlgen/cmd/sqlgen-goderive
sqlgen-goderive $ARG
goimports -w $ARG
