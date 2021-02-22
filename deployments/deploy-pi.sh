#!/bin/sh

sudo apt-get -y install jq

url=`curl https://api.github.com/repos/u-one/go-el-controller/releases/latest | jq '.assets[] | select(.name == "elexporter_linux_arm") | .browser_download_url' | sed 's/"//g'`

echo $url 
wget -q $url

chmod +x ./elexporter_linux_arm

DIR=/opt/u-one/el-exporter

mkdir -p $DIR

mv elexporter_linux_arm $DIR/elexporter
mv elexporter.service /etc/systemd/system/

systemctl enable elexporter.service

