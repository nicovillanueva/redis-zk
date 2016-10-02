#!/bin/bash

IMAGENAME='lapp-dvde004:5000/watcher-zk'

CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
# cp /etc/ssl/certs/ca-certificates.crt .
docker build -t $IMAGENAME:$(cat ../version.txt) .
