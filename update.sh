#!/bin/bash

SUB_DOMAIN=$1
BUILD_NUMBER=$2
WORK_DIR=$3

cd $WORK_DIR


echo "Copying build files ..."
gsutil -m cp gs://pos-proxy/$SUB_DOMAIN/$BUILD_NUMBER/pos-proxy .
gsutil -m cp -r gs://pos-proxy/$SUB_DOMAIN/$BUILD_NUMBER/templates .

mkdir -p /usr/local/bin
mkdir -p /var/www/templates/

echo "Replacing binaries .."
chmod +x ./pos-proxy
mv ./pos-proxy /usr/local/bin/pos-proxy
cp -r ./templates/* /var/www/templates

echo "Restarting proxy ..."
supervisorctl restart all &

