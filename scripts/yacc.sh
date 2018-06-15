#!/bin/bash
set -e
go generate github.com/ng-vu/sqlgen/core/dsl
goimports -w $GOPATH/src/github.com/ng-vu/sqlgen/core/dsl/y.go
