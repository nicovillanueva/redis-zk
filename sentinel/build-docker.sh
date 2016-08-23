#!/bin/bash

IMAGENAME="lapp-dvde004:5000/sentinel-zk:$(cat ../version.txt)"

docker build -t $IMAGENAME .

echo "Built: $IMAGENAME"
