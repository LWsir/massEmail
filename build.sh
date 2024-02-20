#!/bin/bash
echo "注意！:【打包需要GO环境版本 > 1.15.*】"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags 'netcgo',"jsoniter" -o groupEmail -ldflags "-w -s"
chmod +x ./groupEmail