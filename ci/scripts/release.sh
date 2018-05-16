#!/usr/bin/env bash

set -eux

home="$(pwd -P)"

# Auth gcloud
set +x
gcloud auth activate-service-account --key-file <(echo "$GCR_KEY")
set -x

# Configure gcloud
gcloud config set project corded-guild-155314

# Publish build
gsutil -m cp -r "gs://pos-proxy/test/${BUILD_ID}/*" "gs://pos-proxy/${TARGET_ENV}/${BUILD_ID}/"
gsutil -m acl -r ch -u AllUsers:R "gs://pos-proxy/${TARGET_ENV}/${BUILD_ID}/*"
