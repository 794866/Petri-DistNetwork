#!/bin/bash

killall distributed
rm -r Logs

echo "Building project..."
go build distributed.go

go test -v distributed_test.go