#!/usr/bin/env sh

set -eux

CI_PIPELINE_ID="$(cut -d. -f1 ./build_number/version)"
home="$(pwd -P)"

mkdir -p $GOPATH/src/pos-proxy
cd repo
cp -r * $GOPATH/src/pos-proxy
cd $GOPATH/src/pos-proxy
go get
go build -ldflags "-X pos-proxy/config.BuildNumber=${CI_PIPELINE_ID} -X pos-proxy/config.Version=2.0.0"
cp $GOPATH/src/pos-proxy/pos-proxy $home/build
cp $GOPATH/src/pos-proxy/update.sh $home/build
cp -r templates $GOPATH/src/pos-proxy/pos-proxy $home/build
