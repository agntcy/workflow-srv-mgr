FROM workflowserver:latest

ARG AGENT_DIR
ARG MANIFEST_FILE
ARG ENV_FILE

WORKDIR /opt/agent-workflow-server

COPY $AGENT_DIR /opt/agent_src
RUN poetry run pip install /opt/agent_src

#COPY $MANIFEST_FILE /opt/agent-workflow-server/manifest.json
#COPY $ENV_FILE /opt/agent-workflow-server/.env


CMD ["poetry" ,"run", "server"]
