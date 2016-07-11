#!/bin/bash

# optional
msg="Yow. $1"

# in unix seconds
d=`date "+%s"`

input='{
    "service":  "alpha",
    "address":  "1.2.3.4",
    "createdOn": '$d',
    "severity": "Debug",
    "message":  "'"$msg"'"
}'

url="https://pz-logger.$PZDOMAIN/message"
echo
echo POST $url
echo "$input"

echo RETURN:
curl -S -s -XPOST -d "$input" "$url"
echo
