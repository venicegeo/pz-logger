#!/bin/bash

cat > tmp <<foo
{
    "service":  "noservice",
    "address":  "1.2.3.4",
    "createdOn": 123456789,
    "severity": "Debug",
    "message":  "Yow!"
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
