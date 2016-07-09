#!/bin/bash

# should return C, B

url="https://pz-logger.int.geointservices.io/message?perPage=2&page=1"

echo
echo GET $url

ret=$(curl -S -s -XGET "$url")

echo RETURN:
echo "$ret"
echo
