# `spate` Demo

This example producer-consumer application demonstrates the Docker Swarm autoscaler [`spate`](https://github.com/mtneug/spate).

## Quick Start

```sh
$ docker network create -d overlay spate-demo

$ docker service create \
    --name consumer \
    --network spate-demo \
    --replicas 2 \
    mtneug/spate-demo consumer

$ docker service create \
    --name producer \
    --network spate-demo \
    --publish "5000:5000" \
    --mount "type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock" \
    --replicas 1 \
    mtneug/spate-demo producer
```

## License

MIT (C) Matthias Neugebauer
