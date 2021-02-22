#!/bin/sh

sudo apt-get install jq

url=`curl https://api.github.com/repos/u-one/go-el-controller/releases/latest | jq '.assets[] | select(.name == "elexporter_linux_arm") | .browser_download_url' | sed 's/"//g'`

echo $url 
wget $url

pkill elexporter

nohup ./elexporter_linux_arm &

