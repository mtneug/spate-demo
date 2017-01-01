# `spate` Demo

This example producer-consumer application demonstrates the Docker Swarm autoscaler [`spate`](https://github.com/mtneug/spate).

## Quick Start

```sh
$ docker network create -d overlay spate-demo

$ docker service create \
    --name producer \
    --network spate-demo \
    --constraint "node.role == manager" \
    --publish "5000:5000" \
    --mount "type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock" \
    --replicas 1 \
    mtneug/spate-demo producer

$ docker service create \
    --name consumer \
    --network spate-demo \
    --label "de.mtneug.spate=enable" \
    --label "de.mtneug.spate.autoscaler.period=2s" \
    --label "de.mtneug.spate.autoscaler.cooldown.scaled_up=5s" \
    --label "de.mtneug.spate.autoscaler.cooldown.scaled_down=5s" \
    --label "de.mtneug.spate.replicas.min=1" \
    --label "de.mtneug.spate.replicas.max=25" \
    --label "de.mtneug.spate.metric.demo.observer.period=1s" \
    --label "de.mtneug.spate.metric.demo.type=prometheus" \
    --label "de.mtneug.spate.metric.demo.kind=system" \
    --label "de.mtneug.spate.metric.demo.prometheus.endpoint=http://producer:5000/metrics" \
    --label "de.mtneug.spate.metric.demo.prometheus.name=spate_demo_store" \
    --label "de.mtneug.spate.metric.demo.aggregation.method=avg" \
    --label "de.mtneug.spate.metric.demo.aggregation.amount=5" \
    --label "de.mtneug.spate.metric.demo.target=10" \
    --replicas 5 \
    mtneug/spate-demo consumer

$ docker service create \
    --name spate \
    --network spate-demo \
    --constraint "node.role == manager" \
    --mount "type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock" \
    --replicas 1 \
    mtneug/spate
```

## License

MIT (c) Matthias Neugebauer
