#!/bin/bash

rm -r results/*

echo "Building project..."
go build distributed.go

go test -v distributed_test.go