#!/bin/bash

# should return C, B

url="https://pz-logger.int.geointservices.io/message?perPage=2&page=1"

echo
echo GET $url
echo "$json"

ret=$(curl -S -s -XGET -d "$json" "$url")

echo RETURN:
echo "$ret"
echo
