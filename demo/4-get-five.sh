#!/bin/bash

url="https://pz-logger.int.geointservices.io/message?perPage=5"

echo
echo GET $url
echo "$json"

ret=$(curl -S -s -XGET -d "$json" "$url")

echo RETURN:
echo "$ret"
echo
