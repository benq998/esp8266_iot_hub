#!/bin/bash

basepath=$(cd `dirname $0`; pwd)
echo "GOPATH=$basepath"
GOPATH=$basepath go install -v -gcflags "-N -l" ./...

