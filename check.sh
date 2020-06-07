#!/bin/bash

# Run tests
echo "=== Unit tests"
go test -v ./... -test.failfast || exit 1

if [ "$1" == "test" ]; then
    # Done!
    exit 0
fi

# Conform to formatting
echo "=== Code formatting"
gofmt -w -s -l . || exit 1

echo "=== CHECKED"
