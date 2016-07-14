#!/bin/bash

url="https://pz-logger.$PZDOMAIN/message?perPage=1000&sortBy=severity&order=asc"

echo
echo GET $url

echo RETURN:
ret=$(curl -S -s -XGET "$url")

echo "$ret" | grep severity

