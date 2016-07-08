#!/bin/bash

# optional
msg=$1

# in unix ms
d=`date "+%s"`

cat > tmp <<foo
{
    "service":  "alpha",
    "address":  "1.2.3.4",
    "createdOn": $d,
    "severity": "Debug",
    "message":  "Yow! $msg"
}
foo

json=$(cat tmp)

url="https://pz-logger.int.geointservices.io/message"
echo
echo POST $url
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$url")

echo RETURN:
echo "$ret"
echo

rm -f tmp
