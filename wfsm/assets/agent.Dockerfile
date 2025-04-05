# Copyright AGNTCY Contributors (https://github.com/agntcy)
# SPDX-License-Identifier: Apache-2.0

ARG BASE_IMAGE=ghcr.io/agntcy/acp/wfsrv:latest
FROM $BASE_IMAGE

ARG AGENT_DIR
ARG AGENT_FRAMEWORK
ARG AGENT_OBJECT

WORKDIR /opt/agent-workflow-server

COPY $AGENT_DIR /opt/agent_src
RUN poetry run pip install /opt/agent_src

COPY manifest.json /opt/spec/manifest.json
ENV AGENT_MANIFEST_PATH=/opt/spec/manifest.json

COPY start_agws.sh /opt/start_agws.sh
RUN chmod +x /opt/start_agws.sh

ENV AGWS_STORAGE_FILE=/opt/storage/agws_storage.pkl

ENV AGENT_FRAMEWORK=$AGENT_FRAMEWORK
ENV AGENT_OBJECT=$AGENT_OBJECT

ENTRYPOINT ["/opt/start_agws.sh"]
