#!/bin/bash

rm -r results/*
rm -r Logs/*

echo "Building project..."
go build distributed.go

go test -v distributed_test.go