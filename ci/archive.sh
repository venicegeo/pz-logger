#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null

#----------------------------------------------------------------------

sh $root/ci/do_build.sh

#----------------------------------------------------------------------

app=$GOPATH/bin/pz-logger

# gather some data about the repo
source $root/ci/vars.sh

# stage the artifact(s) for a mvn deploy
tar cvzf "$root"/"$APP".tgz \
    $app \
    *.cov \
    lint.txt \
    glide.*
mv $app $root/$APP.$EXT
