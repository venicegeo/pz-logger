#!/bin/bash

d=`date "+%s"`
echo $d

cat > tmp.1 <<foo
{
    "service":  "noservice",
    "address":  "1.2.3.4",
    "createdOn": $d,
    "severity": "Debug",
    "message":  "1111 $d"
}
foo

cat > tmp.2 <<foo
{
    "service":  "noservice",
    "address":  "1.2.3.4",
    "createdOn": $d,
    "severity": "Debug",
    "message":  "2222 $d"
}
foo

cat > tmp.3 <<foo
{
    "service":  "noservice",
    "address":  "1.2.3.4",
    "createdOn": $d,
    "severity": "Debug",
    "message":  "3333 $d"
}
foo

cat > tmp.4 <<foo
{
    "service":  "noservice",
    "address":  "1.2.3.4",
    "createdOn": $d,
    "severity": "Debug",
    "message":  "4444 $d"
}
foo

cat > tmp.5 <<foo
{
    "service":  "noservice",
    "address":  "1.2.3.4",
    "createdOn": $d,
    "severity": "Debug",
    "message":  "5555 $d"
}
foo

json1=$(cat tmp.1)
json2=$(cat tmp.2)
json3=$(cat tmp.3)
json4=$(cat tmp.4)
json5=$(cat tmp.5)

url="https://pz-logger.int.geointservices.io/message"
echo

echo POST $url

ret=$(curl -S -s -XPOST -d "$json1" "$url")

echo RETURN:
echo "$ret"
echo

echo POST $url

ret=$(curl -S -s -XPOST -d "$json2" "$url")

echo RETURN:
echo "$ret"
echo

echo POST $url

ret=$(curl -S -s -XPOST -d "$json3" "$url")

echo RETURN:
echo "$ret"
echo

echo POST $url

ret=$(curl -S -s -XPOST -d "$json4" "$url")

echo RETURN:
echo "$ret"
echo
echo POST $url

ret=$(curl -S -s -XPOST -d "$json5" "$url")

echo RETURN:
echo "$ret"
echo
