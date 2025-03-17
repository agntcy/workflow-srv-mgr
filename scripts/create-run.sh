#!/usr/bin/env bash

curl -X POST -H "Content-Type: application/json" http://0.0.0.0:8000/runs -d @run.json | jq -r .run_id