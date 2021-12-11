#!/bin/bash

rm -r Logs

echo "Building project..."
go build distributed.go

go test -v distributed_test.go