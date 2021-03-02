#!/bin/sh

printenv

echo "start"
echo $BROUTEID
echo $BROUTEPW

## setup
#sudo apt-get -y install jq
#
## get latest binary
#url=`curl https://api.github.com/repos/u-one/go-el-controller/releases/latest | jq '.assets[] | select(.name == "smartmeter-exporter_linux_arm") | .browser_download_url' | sed 's/"//g'`
#echo $url 
#wget -q $url
#chmod +x ./smartmeter-exporter_linux_arm
#
## stop service
#systemctl stop smartmeter-exporter.service
#
## install
#DIR=/opt/u-one/echonetlite
#mkdir -p $DIR
#mv smartmeter-exporter_linux_arm $DIR/smartmeter-exporter
#mv smartmeter-exporter.service /etc/systemd/system/
#
#systemctl enable smartmeter-exporter.service
#
## start service
#systemctl start smartmeter-exporter.service
#
