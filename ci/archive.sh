#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null

#----------------------------------------------------------------------

export GOPATH=$root/gogo
mkdir -p "$GOPATH"

# glide expects these to already exist
mkdir "$GOPATH"/bin "$GOPATH"/src "$GOPATH"/pkg

PATH=$PATH:"$GOPATH"/bin

# build ourself, and go there
go get github.com/venicegeo/pz-logger
cd $GOPATH/src/github.com/venicegeo/pz-logger

#----------------------------------------------------------------------

src=$GOPATH/bin/pz-logger

# gather some data about the repo
source $root/ci/vars.sh

root0="$root"/../0-test

# stage the artifact(s) for a mvn deploy
tar cvzf "$root"/"$APP".tgz \
    $src \
    $root0/*.cov \
    $root0/lint.txt \
    $root0/glide.*
mv $src $root/$APP.$EXT
