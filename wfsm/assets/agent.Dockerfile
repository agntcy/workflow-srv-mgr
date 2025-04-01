# Copyright AGNTCY Contributors (https://github.com/agntcy)
# SPDX-License-Identifier: Apache-2.0

ARG BASE_IMAGE=ghcr.io/agntcy/acp/wfsrv:latest
FROM $BASE_IMAGE

ARG AGENT_DIR

WORKDIR /opt/agent-workflow-server

COPY $AGENT_DIR /opt/agent_src
RUN poetry run pip install /opt/agent_src

ENV API_KEY=$API_KEY
ENV AGENT_MANIFEST_PATH=$AGENT_MANIFEST_PATH
ENV AGENT_FRAMEWORK=$AGENT_FRAMEWORK
ENV AGENTS_REF=$AGENTS_REF

CMD ["poetry" ,"run", "server"]
