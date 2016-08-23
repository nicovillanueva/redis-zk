#!/bin/bash

IMAGENAME="lapp-dvde004:5000/redis-zk:$(cat ../version.txt)"

docker build -t $IMAGENAME .

echo "Built: $IMAGENAME"
