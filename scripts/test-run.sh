#!/usr/bin/env bash

run_id=$(./create-run.sh)
./get-run.sh "$run_id"