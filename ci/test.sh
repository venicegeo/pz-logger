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
go get github.com/venicegeo/pz-gocommon
go get github.com/venicegeo/pz-gocommon/elasticsearch

# ourself
go get github.com/venicegeo/pz-logger/lib
go test -v -coverprofile=logger.cov -coverpkg github.com/venicegeo/pz-logger/lib github.com/venicegeo/pz-logger/lib/tests

###
