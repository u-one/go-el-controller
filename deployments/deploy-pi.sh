#!/bin/sh

# setup
sudo apt-get -y install jq

# get latest binary
url=`curl https://api.github.com/repos/u-one/go-el-controller/releases/latest | jq '.assets[] | select(.name == "elexporter_linux_arm") | .browser_download_url' | sed 's/"//g'`
echo $url 
wget -q $url
chmod +x ./elexporter_linux_arm

# stop service
systemctl stop el-exporter.service

# install
DIR=/opt/u-one/el-exporter
mkdir -p $DIR
mv elexporter_linux_arm $DIR/elexporter
mv el-exporter.service /etc/systemd/system/

systemctl enable el-exporter.service

# start service
systemctl start el-exporter.service

