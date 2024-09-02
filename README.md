# jilo-agent

## overview

Jilo Agent - a remote agent for Jilo Web

Initial version is in PHP.

The current version is in "go" folder and is in Go.

## license

This project is licensed under the GNU General Public License version 2 (GPL-2.0). See LICENSE file.

## installation

Clone the git repo. Either run the agent with Goor build it and run the executable.

Run it (mainly used for tests):

```bash
go run main.go
```

Build the agent:

```bash
go build -o jilo-agent main.go
```

## configuration

The config file is "jilo-agent.json", in the same folder as the "jilo-agent" binary.

You can run the agent without a config file - then default vales are used.

## usage

Run the agent

```bash
./jilo-agent
```

Send queries to its port (by default 8081, in order to avoid 80, 8080, 8888; configurable in jilo-agent.json):

```bash
curl -s http://localhost:8081/nginx
curl -s http://localhost:8081/jicofo
etc...
```
