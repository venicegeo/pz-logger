#!/bin/bash

url="https://pz-logger.$PZDOMAIN/message?perPage=5"

echo
echo GET $url

echo RETURN:
curl -S -s -XGET $url
echo
