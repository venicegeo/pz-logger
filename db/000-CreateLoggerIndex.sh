#!/bin/bash
INDEX_NAME=pzlogger5
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
			"hostName": { "index": "not_analyzed", "type": "string" },
			"application": { "index": "not_analyzed", "type": "string" },
			"process": { "index": "not_analyzed", "type": "string" },
			"messageId": { "index": "not_analyzed", "type": "string" },
			"auditData": {
				"dynamic": "strict",
				"properties": {
					"actor": { "index": "not_analyzed", "type": "string" },
					"actee": { "index": "not_analyzed", "type": "string" },
					"action": { "index": "not_analyzed", "type": "string" },
					"request": { "index": "not_analyzed", "type": "string" }
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
	}'
IndexSettings="
{
	"\""mappings"\"": {
		$LogMapping	
	}
}"


bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP "$IndexSettings" "$LogMapping" $TESTING
