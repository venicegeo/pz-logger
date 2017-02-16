#!/bin/bash
INDEX_NAME=pzlogger5
ALIAS_NAME=$1
ES_IP=$2
TESTING=true
Schema='
		"LogData": {
			"dynamic": "strict",
			"properties": {
				"facility": { "type": "integer" },
				"severity": { "type": "integer" },
				"version": { "type": "integer" },
				"timeStamp": {
					"type": "date",
					"format": "yyyy-MM-dd'\''T'\''HH:mm:ssZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSSSSZZ"					
				},
				"hostName": { "index": "not_analyzed", "type": "string" },
				"application": { "index": "not_analyzed", "type": "string" },
				"process": { "index": "not_analyzed", "type": "string" },
				"messageId": { "index": "not_analyzed", "type": "string" },
				"auditData": {
					"dynamic": "strict",
					"properties": {
						"actor": { "index": "not_analyzed", "type": "string" },
						"actee": { "index": "not_analyzed", "type": "string" },
						"action": { "index": "not_analyzed", "type": "string" }
					}
				},
				"metricData": {
					"dynamic": "strict",
					"properties": {
						"name": { "index": "not_analyzed", "type": "string" },
						"value": { "type": "double" },
						"object": { "index": "not_analyzed", "type": "string" }
					}
				},
				"sourceData": {
					"dynamic": "strict",
					"properties": {
						"file": { "index": "not_analyzed", "type": "string" },
						"line": { "type": "integer" },
						"function": { "index": "not_analyzed", "type": "string" }
					}
				},
				"message": { "index": "not_analyzed", "type": "string" }
			}
		}
	'

if [[ $ALIAS_NAME == "" ]]; then
  echo "Please specify an alias name as argument 1"
  exit 1
fi

if [[ $ES_IP == "" ]]; then
  echo "Please specify the elasticsearch ip as argument 2"
  exit 1
fi 

if [[ $ES_IP != */ ]]; then
  ES_IP="$ES_IP/"
fi

#
# Check to see if index already exists
#

echo "Checking to see if index $INDEX_NAME already exists..."
cat=_cat
catCurl=`curl -X GET -H "Content-Type: application/json" -H "Cache-Control: no-cache" "$ES_IP$cat/indices" --write-out %{http_code} 2>/dev/null`
http_code=`echo $catCurl | cut -d] -f2`
if [[ "$http_code" != 200 ]]; then
  echo "Status code $http_code returned while checking indices"
  exit 1
fi

if [[ $catCurl == *""\""index"\"":"\""$INDEX_NAME"\"""* ]]; then
  echo "Index $INDEX_NAME already exists"
  exit 0
fi

#
# Create the index
#

echo "Creating index $INDEX_NAME with mappings..."
createIndexCurl=`curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{ "\""mappings"\"": {$Schema}}" "$ES_IP$INDEX_NAME" --write-out %{http_code} 2>/dev/null`
echo $createIndexCurl
http_code=`echo $catCurl | cut -d] -f2`
if [[ $createIndexCurl != '{"acknowledged":true}200' ]]; then
  echo "Failed to create index $INDEX_NAME. Code: $http_code"
  exit 1
fi

aliases=_aliases

#
# If testing, create two indices that have the alias we are trying to set
#

if [ "$TESTING" = true ] ; then
    echo "Creating test indices..."
    apple=apple
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{}" "$ES_IP$apple" --write-out %{http_code}; echo " "
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{
    "\""actions"\"" : [ { "\""add"\"" : { "\""index"\"" : "\""apple"\"", "\""alias"\"" : "\""$ALIAS_NAME"\"" } } ]
    }" "$ES_IP$aliases" --write-out %{http_code}; echo " "
    pear=pear
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{}" "$ES_IP$pear" --write-out %{http_code}; echo " "
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{
    "\""actions"\"" : [ { "\""add"\"" : { "\""index"\"" : "\""pear"\"", "\""alias"\"" : "\""$ALIAS_NAME"\"" } } ]
    }" "$ES_IP$aliases" --write-out %{http_code}; echo " "
fi

#
# Search for indices that are using the alias we are trying to set
#

getAliasesCurl=`curl -XGET -H "Content-Type: application/json" -H "Cache-Control: no-cache" "$ES_IP$cat/aliases" --write-out %{http_code} 2>/dev/null`
http_code=`echo $catCurl | cut -d] -f2`
if [[ "$http_code" != 200 ]]; then
  echo "Status code $http_code returned from catting aliases"
  exit 1
fi

#
# Extract index names that are using the alias from the above result
#

regex=""\""alias"\"":"\""$ALIAS_NAME"\"","\""index"\"":"\""([^"\""]+)"
temp=`echo $getAliasesCurl|grep -Eo $regex | cut -d\" -f8`
indexArr=(${temp// / })
echo "Found ${#indexArr[@]} indices currently using alias $ALIAS_NAME: ${indexArr[@]}"

#
# Remove alias from all above indices
#

for index in ${indexArr[@]}
do
  removeAliasCurl=`curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{
    "\""actions"\"" : [
        { "\""remove"\"" : { "\""index"\"" : "\""$index"\"", "\""alias"\"" : "\""$ALIAS_NAME"\"" } }
    ]
  }" "$ES_IP$aliases" --write-out %{http_code} 2>/dev/null`
  http_code=`echo $catCurl | cut -d] -f2`
  if [[ $removeAliasCurl != '{"acknowledged":true}200' ]]; then
    echo "Failed to remove alias $ALIAS_NAME on index $index. Code: $http_code"
    exit 1
  fi
  echo "Removed alias $ALIAS_NAME on index $index"
done

#
# Create alias on our index
#

createAliasCurl=`curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{
    "\""actions"\"" : [
        { "\""add"\"" : { "\""index"\"" : "\""$INDEX_NAME"\"", "\""alias"\"" : "\""$ALIAS_NAME"\"" } }
    ]
}" "$ES_IP$aliases" --write-out %{http_code} 2>/dev/null`
http_code=`echo $catCurl | cut -d] -f2`
if [[ $createIndexCurl != '{"acknowledged":true}200' ]]; then
  echo "Failed to create alias $ALIAS_NAME on index $INDEX_NAME. Code: $http_code"
  exit 1
fi
echo "Created alias $ALIAS_NAME on index $INDEX_NAME"

echo 
echo "Success!"
