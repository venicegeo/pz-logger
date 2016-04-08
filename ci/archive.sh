#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null

export GOPATH=$root/gogo
mkdir -p $GOPATH

###

go get github.com/venicegeo/pz-logger
go get github.com/venicegeo/pz-logger/server

go install github.com/venicegeo/pz-logger

###

src=$GOPATH/bin/pz-logger

# gather some data about the repo
source $root/ci/vars.sh

# stage the artifact for a mvn deploy
mv $src $root/$APP.$EXT
