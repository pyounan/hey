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
  apt-get update && apt-get install -y libssl-dev supervisor wget mongodb-server openssl curl google-cloud-sdk redis-server
}

pull_proxy(){
  # fetch the proxy program (binary file and templates)
  AVAILABLE_VERSIONS=($(gsutil ls gs://pos-proxy/$SUB_DOMAIN/ | cut -d"/" -f 5 | sort -nr))

  gsutil -m cp gs://pos-proxy/$SUB_DOMAIN/${AVAILABLE_VERSIONS[0]}/pos-proxy .
  gsutil -m cp -r gs://pos-proxy/$SUB_DOMAIN/${AVAILABLE_VERSIONS[0]}/templates .

  mkdir -p /usr/local/bin
  sudo supervisorctl stop all || true
  cp ./pos-proxy /usr/local/bin/pos-proxy
  mkdir -p /var/www/templates/
  cp -r ./templates/* /var/www/templates

  chmod +x /usr/local/bin/pos-proxy
}

write_config(){
# create supervisor configuration file
FILE=/etc/supervisor/conf.d/pos_proxy.conf
touch $FILE
cat <<EOM >$FILE
[program:pos-proxy]
command=/usr/local/bin/pos-proxy --templates="/var/www/templates/*"
autostart=true
autorestart=true
stderr_logfile=/var/log/pos_proxy.err.log
stdout_logfile=/var/log/pos_proxy.out.log
EOM

# copy the configuration file to/etc/cloudinn/pos_config.json
mkdir -p /etc/cloudinn || true
# make file for auth credentials
FILE=/etc/cloudinn/auth_credentials
touch $FILE
cat <<EOM >$FILE
$AUTH_USERNAME,$AUTH_PASSWORD
EOM
FILE=/etc/cloudinn/pos_config.json
touch $FILE
curl -u $AUTH_USERNAME:$AUTH_PASSWORD https://$SUB_DOMAIN.cloudinn.net/api/pos/proxy/settings/ -o /etc/cloudinn/pos_config.json
}

if [ "$(id -u)" != "0" ]; then
    error "This script should be run using sudo or as the root user"
    exit 1
fi

if [ "$1" == "" ]; then
	error "Please provide auth username"
	exit
fi

if [ "$2" == "" ]; then
	error "Please provide auth password"
	exit
fi

if [ "$3" == "" ]; then
	error "Please provide cloudinn subdomain"
	exit
fi

AUTH_USERNAME=$2
AUTH_PASSWORD=$3
SUB_DOMAIN=$4

apt-get install curl
add_google_repo
install_deps
pull_proxy
write_config
sudo systemctl restart mongodb.service
eudo systemctl restart supervisor.service
