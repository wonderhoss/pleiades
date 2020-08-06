# Pleiades - A WMF Event Stream Relay

* _Note: Pleiades is a demonstration project, provided as-is_

<img align="left" width="" height="" src="https://raw.githubusercontent.com/gargath/pleiades/mainline/pleiades_logo.png" alt="Pleiades Logo">
Pleiades subscribes to the `recentchange` event stream [provided by the Wikimedia Foundation](https://wikitech.wikimedia.org/wiki/Event_Platform/EventStreams) and re-publishes each event
received to either a Kafka topic or a separate file on the filesystem.

It supports resuming subscriptions from historic event IDs in case of interruption.

Prometheus metrics are provided on `/metrics`.

## Usage

```
$ pleiades --help
Usage of pleiades:
      --file.enable              enable the file publisher
      --file.publishDir string   the directory to publish events to (default "./events")
      --help                     print this help and exit
      --kafka.broker string      the kafka broker to connect to (default "localhost:9092")
      --kafka.enable             enable the kafka publisher
      --kafka.topic string       the kafka topic to publish to (default "pleiades-events")
      --metricsPort string       the port to serve Prometheus metrics on (default "9000")
  -q, --quiet                    quiet output - only show ERROR and above
  -r, --resume                   try to resume from last seen event ID (default true)
  -v, --verbose                  enable verbose output
  ```

* Only one publisher can be enabled at a time, use either `--file.enable` or `--kafka.enable`.
* `--metricsPort` sets the port to use for the Prometheus metrics endpoint (see below)
* `--kafka.broker` and `--kafka.topic` set the broker and topic to publish do when using Kafka
  Please note that currently only one single broker and single-partition topic is supported
* When using the file publisher, `--file.publishDir` sets the directory on the filesystem to store events
  If it does not exist, it will be created
* `-q` and `-v` are mutually exclusive and decrease or increase the log level respectively
* Setting `-r=false` will disable the subscription resume mechanism and start consuming events from the current point in time


## Metrics

Pleiades exposes the following metrics in addition to the standard go runtime stats:

| name | type | help |
|------|------|------|
| `pleiades_recv_events_total` | counter | Total number of parsed events recenved from upstream |
| `pleiades_recv_event_lines_total` | counter | Total number of raw lines read from upstream, regardless of whether they become part of an event object |
| `pleiades_recv_errors_total` | counter | Total number of errors encountered by the consumer |
| `pleiades_goroutine_restarts` | counter | Number of times any of the interal goroutines restarted after encountering an error |
| `pleiades_file_publish_events_total` | counter | Total number of events written to file by the filesystem publisher |
| `pleiades_file_publish_errors_total` | counter | Total number of errors encountered while writing to filesystem - each is likely to have dropped one event |
| `pleiades_kafka_publish_events_total` | counter | Total number of events published to Kafka |
| `pleiades_kafka_publish_writes_total` | counter | Total number of write operations published to Kafka |
| `pleiades_kafka_writer_errors_total` | counter | Total number of errors encountered while publishing to Kafka - each is likely to have dropped one event |
| `pleiades_kafka_publish_write_time_seconds` | gauge | Time spent writing to Kafka ('min', 'max', 'avg') |
| `pleiades_kafka_publish_wait_time_seconds` | gauge | Time spent waiting for Kafka responses ('min', 'max', 'avg') |
| `pleiades_kafka_publish_lag_milliseconds` | gauge | Time difference between receiving an event from upstream and publishing to Kafka |

## Running in KIND

[Kind](https://kind.sigs.k8s.io/) (or kubernetes-in-docker) is a tool for standing up kubernetes clusters on your local machine for testing purposes.

The `/deploy/kind` folder contains all relevant descriptors:

* [kind-cluster.yaml](deploy/kind/kind-cluster.yaml) is a  Kind cluster descriptor including host port forwarding for ports `80` and `443` for use with an Ingress Controller
* [strimzi/strimzi.yaml](deploy/kind/strimzi/strimzi.yaml) deploys the [Strimzi Kafka Operator](https://strimzi.io/)
* [prometheus/](deploy/kind/prometheus) contains descriptors for the [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator), an Ingress and a single Prometheus instance
* [nginx-ingress/](deploy/kind/nginx-ingress/nginx-ingress.yaml) deploys the [Nginx Ingress Controller](https://github.com/kubernetes/ingress-nginx/)
* [pleiades/](deploy/kind/pleiades) contains descriptors for Pleiades itself, the managed Kafka cluster and topic it publishes to, as well as a `ServiceMonitor`

### Putting it all together

First, create the Kind cluster
```
$ pwd
pleiades/deploy/kind

$ kind create cluster --config=kind-cluster.yaml

$ kubectl config use-context kind-kind
```

Deploy the Ingress Controller
```
$ kubectl apply -f nginx-ingress/nginx-ingress.yaml
```
and wait for the pods to become ready.

Next, deploy the Strimzi and Prometheus operators
```
$ kubectl apply -f strimzi/strimzi.yaml
$ kubectl apply -f prometheus/*
```

Again wait for Pods to become ready before deploying the Kafka resources
```
$ kubectl apply -f pleiades/kafka-persistent-single.yaml -f kafkatopic.yaml
```

Once Kafka is running (again, check Pod readiness), deploy Pleiades
```
$ kubectl apply -f pleiades/pleiades-*
```

You should be able to access Prometheus on http://localhost/prometheus and verify that it is scraping Pleiades by checking the `Targets` section.