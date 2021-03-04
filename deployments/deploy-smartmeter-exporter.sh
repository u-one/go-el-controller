#!/bin/sh

export BROUTEID=$1
export BROUTEPW=$2

# setup
sudo apt-get -y install jq

# get latest binary
url=`curl https://api.github.com/repos/u-one/go-el-controller/releases/latest | jq '.assets[] | select(.name == "smartmeter-exporter_linux_arm") | .browser_download_url' | sed 's/"//g'`
echo $url 
wget -q $url
chmod +x ./smartmeter-exporter_linux_arm

# stop service
systemctl stop smartmeter-exporter.service

# install
DIR=/opt/u-one/echonetlite
mkdir -p $DIR
mv smartmeter-exporter_linux_arm $DIR/smartmeter-exporter
mv smartmeter-exporter.service /etc/systemd/system/

cat << EOS > $DIR/start.sh
#!/bin/sh

/opt/u-one/echonetlite/smartmeter-exporter --brouteid=${BROUTEID} --broutepw=${BROUTEPW}
EOS

chmod 755 $DIR/start.sh

systemctl enable smartmeter-exporter.service

# start service
systemctl start smartmeter-exporter.service

