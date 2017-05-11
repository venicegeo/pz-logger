#!/bin/bash
INDEX_NAME=pzlogger6
ALIAS_NAME=$1
ES_IP=$2
TESTING=$3

LogMapping='
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
			"hostName": { "type": "keyword" },
			"application": { "type": "keyword" },
			"process": { "type": "keyword" },
			"messageId": { "type": "keyword" },
			"auditData": {
				"dynamic": "strict",
				"properties": {
					"actor": { "type": "keyword" },
					"actee": { "type": "keyword" },
					"action": { "type": "keyword" }
				}
			},
			"metricData": {
				"dynamic": "strict",
				"properties": {
					"name": { "type": "keyword" },
					"value": { "type": "double" },
					"object": { "type": "keyword" }
				}
			},
			"sourceData": {
				"dynamic": "strict",
				"properties": {
					"file": { "type": "keyword" },
					"line": { "type": "integer" },
					"function": { "type": "keyword" }
				}
			},
			"message": { "type": "text" }
		}
	}'
IndexSettings="
{
	"\""mappings"\"": {
		$LogMapping	
	}
}"


bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP "$IndexSettings" "$LogMapping" $TESTING
