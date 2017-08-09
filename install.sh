#!/bin/bash
error() {
  printf '\E[31m'; echo "$@"; printf '\E[0m'
}

info() {
  printf '\e[34m'; echo "$@"; printf '\E[0m'
}


add_google_repo(){
  # Create an environment variable for the correct distribution
  export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)"
  
  # Add the Cloud SDK distribution URI as a package source
  rm /etc/apt/sources.list.d/google-cloud-sdk.list || true
  
  echo "deb https://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
  
  # Import the Google Cloud Platform public key
  curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
}

install_deps(){
  apt-get update && apt-get install -y libssl-dev supervisor wget mongodb-server openssl curl google-cloud-sdk 
}

pull_proxy(){
  # fetch the proxy program (binary file)
  gcloud auth activate-service-account --key-file ${GS_KEY}
  
  gsutil -m cp gs://pos-proxy/$SUB_DOMAIN/bin/pos-proxy .
  
  mkdir -p /usr/local/bin
  cp ./pos-proxy /usr/local/bin/pos-proxy
  
  chmod +x /usr/local/bin/pos-proxy
}

write_config(){
# create supervisor configuration file
sudo supervistorctl stop all
FILE=/etc/supervisor/conf.d/pos_proxy.conf
touch $FILE
cat <<EOM >$FILE
[program:pos-proxy]
command=/usr/local/bin/pos-proxy -server_crt=/usr/local/certs/server.crt -server_key=/usr/local/certs/server.key
autostart=true
autorestart=true
stderr_logfile=/var/log/pos_proxy.err.log
stdout_logfile=/var/log/pos_proxy.out.log
EOM

# copy the configuration file to/etc/cloudinn/pos_config.json
mkdir -p /etc/cloudinn || true
FILE=/etc/cloudinn/pos_config.json
curl -o $FILE -H "Authorization: JWT $PROXY_TOKEN" https://$SUB_DOMAIN.cloudinn.net/api/pos/proxy/settings/
# make file for proxy token
FILE=/etc/cloudinn/proxy_token.json
touch $FILE
cat <<EOM >$FILE
{
	"proxy_token": "$PROXY_TOKEN"
}
EOM
}

if [ "$(id -u)" != "0" ]; then
    error "This script should be run using sudo or as the root user"
    exit 1
fi

if [ "$1" == "" ]; then
    error "Please provide google cloud access key"
    exit 1
fi

if [ "$2" == "" ]; then
	error "Please provide proxy token"
	exit
fi

if [ "$3" == "" ]; then
	error "Please provide cloudinn subdomain"
	exit
fi

if [ ! -f "$1" ]; then
    error "Key not found!"
    exit 1
fi

GS_KEY=$1
PROXY_TOKEN=$2
SUB_DOMAIN=$3

apt-get install curl
add_google_repo
install_deps
pull_proxy
write_config
sudo systemctl restart mongodb.service
sudo systemctl restart supervisor.service
