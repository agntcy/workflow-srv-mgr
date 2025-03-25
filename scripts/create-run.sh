#!/usr/bin/env bash

curl -X POST -H "Content-Type: application/json" -H "x-api-key: $API_KEY" http://127.0.0.1:$PORT/runs -d @run.json | jq -r .run_id