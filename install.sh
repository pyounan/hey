#!/bin/bash
error() {
  printf '\E[31m'; echo "$@"; printf '\E[0m'
}

info() {
  printf '\e[34m'; echo "$@"; printf '\E[0m'
}

#if [[ $EUID -eq 0 ]]; then
if [ "$(id -u)" != "0" ]; then
    error "This script should be run using sudo or as the root user"
    exit 1
fi

if [ "$1" == "" ]; then
    error "Please provide google cloud access key"
    exit 1
fi

if [ ! -f "$1" ]; then
    error "Key not found!"
    exit 1
fi

mkdir -p /usr/local/certs/
if [ ! -f /usr/local/certs/server.crt ]; then
    info "Generating self signed certificate ..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout /usr/local/certs/server.key -out /usr/local/certs/server.crt
    info "Certificate created!"
fi

GS_KEY=$1

# install mongodb
apt-get install -y libssl-dev python-pip supervisor wget mongodb-server openssl curl

service supervisor restart

#service mongodb enable
service mongodb start

# install gsutil
pip install pyopenssl
# Create an environment variable for the correct distribution
export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)"

# Add the Cloud SDK distribution URI as a package source
rm /etc/apt/sources.list.d/google-cloud-sdk.list || true

echo "deb https://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list

# Import the Google Cloud Platform public key
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -

# Update the package list and install the Cloud SDK
apt-get update && apt-get install -y google-cloud-sdk

# fetch the proxy program (binary file)
gcloud auth activate-service-account --key-file ${GS_KEY}

gsutil -m cp gs://pos-proxy/staging/bin/pos-proxy .
pwd

mkdir -p /usr/local/bin
cp ./pos-proxy /usr/local/bin/pos-proxy

chmod +x /usr/local/bin/pos-proxy

# create supervisor configuration file
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
touch $FILE
cat <<EOM >$FILE
{
	"backend_uri": "https://staging.cloudinn.net",
    "tenant_id": "4",
    "fdms": [
       {
             "fdm_port": "/dev/ttyS0",
             "fdm_speed": 19200
       }
     ],
     "fdm_mapping": [
        {
             "rcrs": "12345678901234",
             "fdm": "/dev/ttyS0"
        }
     ]
}
EOM
sudo systemctl restart supervisor.service
