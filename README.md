# :octopus: Octopinger

[![Release](https://github.com/ionos-cloud/octopinger/actions/workflows/release.yml/badge.svg)](https://github.com/ionos-cloud/octopinger/actions/workflows/release.yml)
[![Taylor Swift](https://img.shields.io/badge/secured%20by-taylor%20swift-brightgreen.svg)](https://twitter.com/SwiftOnSecurity)
[![Volkswagen](https://auchenberg.github.io/volkswagen/volkswargen_ci.svg?v=1)](https://github.com/auchenberg/volkswagen)

Octopinger is an Kubernetes Operator to monitor the connectivity of your cluster. The probes use `ICMP` to determine the connectivity between cluster nodes. Metrics are exported via [Prometheus](https://prometheus.io/).

## Get Started

The operator is creating a `DeamonSet` to schedula an `octopinger` instance on all Kubernetes nodes. The `octopinger` instances get created with a `ConfigMap` that contains the current running nodes and configuration options. The `ConfigMap` is updated as instances are in the `running` phase and have an IP address assigned.

## Install

Create a namespace for Octopinger.

```bash
kubectl create namespace octopinger
```

Next, install the custom resource defintions, service accounts, roles and operator.

```bash
kubectl apply -n octopinger -f https://raw.githubusercontent.com/ionos-cloud/octopinger/v0.0.36/manifests/install.yaml
```

Now, you are ready to install octopinger to your cluster.

```bash
kubectl apply -n octopinger -f examples/octopinger_simple.yaml
```

## Helm

[Helm](https://helm.sh/) can be used to install :octopus: Octopinger to your cluster.

```bash
helm repo add octopinger https://octopinger.io/
helm repo update 
```

Install Octopinger to your cluster in a `octopinger` namespace.

```bash
helm install octopinger octopinger/octopinger --create-namespace --namespace octopinger
```

## Metrics

This is the list of Prometheus metrics :octopus: Octopinger is exporting.

* `octopinger_probe_nodes_total`
* `octopinger_probe_nodes_reports`
* `octopinger_probe_rtt_min`
* `octopinger_probe_rtt_mean`
* `octopinger_probe_rtt_max`
* `octopinger_probe_loss_min`
* `octopinger_probe_loss_max`
* `octopinger_probe_loss_mean`
* `octopinger_probe_health_total`

## License

[Apache 2.0](/LICENSE)
