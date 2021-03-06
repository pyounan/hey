#!/usr/bin/env bash

set -eux

export CI_PIPELINE_ID="$(cut -d. -f1 ./proxy_build_number/version)"
home="$(pwd -P)"

# Auth gcloud
set +x
gcloud auth activate-service-account --key-file <(echo "$GCR_KEY")
set -x

# Configure gcloud
gcloud config set project corded-guild-155314

# Publish build
gsutil -m cp -r build/* gs://pos-proxy/test/${CI_PIPELINE_ID}/
gsutil -m acl -r ch -u AllUsers:R gs://pos-proxy/test/${CI_PIPELINE_ID}/*
