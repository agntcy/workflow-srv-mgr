#!/usr/bin/env bash

curl -X POST -H "Content-Type: application/json" http://127.0.0.1:$PORT/runs -d @run.json | jq -r .run_id