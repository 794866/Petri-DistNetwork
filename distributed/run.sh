#!/bin/bash

TestCases="2SubNets2Br 3SubNets2Br 6SubNets5BrHomogen 6SubNets5Br1BrSlow 6SubNets5BrLA"

rm -r results/*
echo "Building project..."
go build distributed.go

for i in ${TestCases}; do
  echo "Running test ${i} ->->->->"
  go test -v  -run "$i" distributed_test.go
done