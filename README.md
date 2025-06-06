# Oxy PubSub Broker

## Description
A proof of concept broker for messaging over TCP implemented in Go.

## Features

- Separate ports for publishers and subscriber (default: `8080` for subscribing, `8081` for publishing)
- Messages are published to all subscribers
- Publishers are notified when a subscriber joins or if there are no subscribers left
- Messages are not persistent
- Messages are delivered asynchronously, any high latency subscriber are not downgrading performance of the broker

## Possible future improvements:

- Persistent messages (external service)
- Topics for reduced unnecessary message traffic for subscribers
- Expanded metadata for messages to allow filtering of messages and easier upkeep
- CLI/WEB UI interface for more user-friendly usage of the broker

## Quickstart

1. Navigate to `docker` directory.
2. Run the following `compose` command to build image and deploy the server:

```sh
docker compose up [-d]
```

`docker-compose.yml` file can be modified to designate which ports should be mapped for subscriber and publishers.

3. Connect to the server over TCP, e.g.:

```sh
telnet localhost 8081
```

4. Publisher connections can publish with the following command:

```
PUB My test message
```