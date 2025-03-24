#!/usr/bin/env bash

curl -H "x-api-key: $API_KEY" http://0.0.0.0:$PORT/runs/"$1"/output