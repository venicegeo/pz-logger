#!/bin/bash

url="https://pz-logger.int.geointservices.io/message?perPage=10"
echo
echo GET $url
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$url")

echo RETURN:
echo "$ret"
echo
