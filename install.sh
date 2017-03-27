# install mongodb
sudo apt-get install -y libssl-dev python-pip supervisor wget mongodb-org openssl

sudo service supervisor restart

sudo service mongod enable
sudo service mongod start

# install gsutil
pip install pyopenssl
# Create an environment variable for the correct distribution
export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)"

# Add the Cloud SDK distribution URI as a package source
echo "deb https://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list

# Import the Google Cloud Platform public key
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -

# Update the package list and install the Cloud SDK
sudo apt-get update && sudo apt-get install -y google-cloud-sdk

# fetch the proxy program (binary file)
gcloud auth activate-service-account --key-file gc_credentials.json

gsutil -m cp gs://pos-proxy/staging/bin/pos-proxy .

cp pox_proxy /usr/local/bin/pos_proxy

sudo chmod +x /usr/local/bin/pos_proxy

# create supervisor configuration file
sudo touch /etc/supervisor/conf.d/pos_proxy.conf
read -r -d '' CONF <<- EOM
[program:long_script]
command=/usr/local/bin/pos_proxy.sh
autostart=true
autorestart=true
stderr_logfile=/var/log/pos_proxy.err.log
stdout_logfile=/var/log/pos_proxy.out.log
EOM

echo "$CONF" > /etc/supervisor/conf.d/pos_proxy.conf

# copy the configuration file to/etc/cloudinn/pos_config.json
sudo mkdir /etc/cloudinn
sudo touch /etc/cloudinn/pos_config.json
read -r -d '' pos_config <<- EOM
{
	"backend_uri": "https://staging.cloudinn.net",
    "tenant_id": 4,
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
sudo echo "$pos_config" > /etc/cloudinn/pos_config.json
