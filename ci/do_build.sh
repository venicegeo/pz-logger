#!/bin/bash -ex

export GOPATH=$root/gogo
mkdir -p "$GOPATH"

# glide expects these to already exist
mkdir "$GOPATH"/bin "$GOPATH"/src "$GOPATH"/pkg

PATH=$PATH:"$GOPATH"/bin

go version

# install metalinter
go get -u github.com/alecthomas/gometalinter
gometalinter --install

# get ourself, and go there
go get github.com/venicegeo/pz-logger
cd $GOPATH/src/github.com/venicegeo/pz-logger

# run unit tests w/ coverage collection
go test -v -coverprofile=logger.cov -coverpkg github.com/venicegeo/pz-logger/logger github.com/venicegeo/pz-logger/logger

# lint
sh ci/metalinter.sh | tee lint.txt
wc -l lint.txt
