# :octopus: Octopinger

[![Test & Build](https://github.com/ionos-cloud/octopinger/actions/workflows/main.yml/badge.svg)](https://github.com/ionos-cloud/octopinger/actions/workflows/main.yml)
[![Taylor Swift](https://img.shields.io/badge/secured%20by-taylor%20swift-brightgreen.svg)](https://twitter.com/SwiftOnSecurity)
[![Volkswagen](https://auchenberg.github.io/volkswagen/volkswargen_ci.svg?v=1)](https://github.com/auchenberg/volkswagen)

Octopinger is an Kubernetes Operator to monitor the connectivity of your cluster. It supports `ICMP` and `TCP` to probe connectivity between cluster and also external references. Metrics are exported via [Prometheus](https://prometheus.io/).

## Get Started

The operator is creating a `DeamonSet` to schedula an `octopinger` instance on all Kubernetes nodes. The `octopinger` instances get created with a `ConfigMap` that contains the current running nodes and configuration options. The `ConfigMap` is updated as instances are in the `running` phase and have an IP address assigned.

## Install

Create a namespace for Octopinger.

```bash
kubectl create namespace octopinger
```

Next, install the custom resource defintions, service accounts, roles and operator.

```bash
kubectl apply -n octopinger -f https://raw.githubusercontent.com/ionos-cloud/octopinger/v0.0.21/manifests/install.yaml
```

Now, you are ready to install octopinger to your cluster.

```bash
kubectl apply -n octopinger examples/octopinger_simple.yaml
```

## License

[Apache 2.0](/LICENSE)
