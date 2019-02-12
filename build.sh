#!/bin/bash

VERSION=$1

GOARCH=amd64 GOOS=darwin go build -o releases/hans-$VERSION-darwin-amd64
GOARCH=amd64 GOOS=linux go build -o releases/hans-$VERSION-linux-amd64
GOARCH=amd64 GOOS=windows go build -o releases/hans-$VERSION-windows-amd64
