#!/bin/bash

url="https://pz-logger.int.geointservices.io/message?perPage=3"

echo
echo GET $url
echo "$json"

ret=$(curl -S -s -XGET -d "$json" "$url")

echo RETURN:
echo "$ret"
echo
