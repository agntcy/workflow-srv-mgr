#!/usr/bin/env bash
: ${AGENT_ID:?"agent id must be provided"}
export AGENTS_REF="{\"$AGENT_ID\": \"$AGENT_OBJECT\"}"
# Run the Poetry server
poetry run server
