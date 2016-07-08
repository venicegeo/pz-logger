#!/bin/bash

url="https://pz-logger.int.geointservices.io/message?perPage=10&sortBy=createdOn&orderBy=desc"
url="https://pz-logger.int.geointservices.io/message?perPage=6"
echo
echo GET $url
echo "$json"

ret=$(curl -S -s -XGET -d "$json" "$url")

echo RETURN:
echo "$ret"
echo
