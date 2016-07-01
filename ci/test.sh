#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null

export GOPATH=$root/gogo
mkdir -p $GOPATH

###

# external pkgs
go get gopkg.in/olivere/elastic.v3
go get github.com/stretchr/testify/suite
go get github.com/stretchr/testify/assert

# our gocommon pkgs
go get github.com/venicegeo/pz-gocommon/gocommon
go get github.com/venicegeo/pz-gocommon/elasticsearch

# ourself
go get github.com/venicegeo/pz-logger/logger
go test -v -coverprofile=logger.cov -coverpkg github.com/venicegeo/pz-logger/logger github.com/venicegeo/pz-logger/logger

###
