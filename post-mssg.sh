#!/bin/bash

cat > tmp <<foo
{
    "service":  "noservice",
    "address":  "1.2.3.4",
    "stamp":    123456789,
    "severity": "Debug",
    "message":  "Yow!"
}
foo

json=$(cat tmp)

url="https://pz-logger.stage.geointservices.io/v1/messages"
echo
echo POST $url
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$url")

echo RETURN:
echo "$ret"
echo
