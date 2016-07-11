#!/bin/bash

# should return C, B

url="https://pz-logger.$PZDOMAIN/message?perPage=2&page=1"

echo
echo GET $url

ret=$(curl -S -s -XGET "$url")

echo RETURN:
echo "$ret"
echo
