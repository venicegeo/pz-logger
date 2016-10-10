#!/bin/bash -ex

pushd `dirname $0`/.. > /dev/null
root=$(pwd -P)
popd > /dev/null
export GOPATH=$root/gogo

#----------------------------------------------------------------------
pwd
sh $root/ci/do_build.sh
pwd
#----------------------------------------------------------------------

app=$GOPATH/bin/pz-logger

# gather some data about the repo
source $root/ci/vars.sh

# stage the artifact(s) for a mvn deploy
ls
ls $root
tar cvzf $root/$APP.tgz \
    $app \
    $root/logger.cov \
    $root/lint.txt \
    $root/glide.lock \
    $root/glide.yaml
ls $root
mv $app $root/$APP.$EXT
