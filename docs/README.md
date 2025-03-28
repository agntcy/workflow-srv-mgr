
## Getting Started

To get a local copy up and running follow these simple steps.

### Prerequisites

Docker engine has to be present on the host running the Workflow Server Manager.


### Installation

Download and unpack the executable binary from the [releases page](https://github.com/agntcy/workflow-srv-mgr/releases)

Alternatively you can execute the installer script by running the following command:
```bash
curl -L https://raw.githubusercontent.com/agntcy/workflow-srv-mgr/refs/heads/install/install.sh | bash
```
The installer script will download the latest release and unpack it into the `bin` folder in the current directory.
The output of the execution looks like this:

```bash
 $ curl -L https://raw.githubusercontent.com/agntcy/workflow-srv-mgr/refs/heads/install/install.sh | bash                                                                                      [17:17:28]
% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   858  100   858    0     0   8524      0 --:--:-- --:--:-- --:--:--  8580
Installing the Workflow Server Manager tool:
OS:	darwin
ARCH:	amd64
TAG:	0.0.1-dev.23
ARCHIVE_URL:	https://github.com/agntcy/workflow-srv-mgr/releases/download/v0.0.1-dev.23/wfsm0.0.1-dev.23_darwin_amd64.tar.gz
Installation complete. The 'wfsm' binary is located at /Users/john/prj/agntcy/workflow-srv-mgr/bin/wfsm```

Listed variables can be overridden by providing the values as environment variables
```


### Usage

Available commands can be listed using the installed tool:

```bash
./wfsm                                                                                                                                                                                                   [17:19:39]

ACP Workflow Server Manager Tool

Wraps an agent into a web server and exposes the agent functionality through ACP.
It also provides commands for managing existing deployments and cleanup tasks

Usage:
  wfsm [command]

Available Commands:
  check       Checks the prerequisites for the command
  completion  Generate the autocompletion script for the specified shell
  deploy      Build an ACP agent
  help        Help about any command
  list        List an ACP agents running in the deployment
  logs        Show logs of an ACP agent deployment(s)
  stop        Stop an ACP agent deployment

Flags:
  -h, --help      help for wfsm
  -v, --version   version for wfsm

Use "wfsm [command] --help" for more information about a command.
```