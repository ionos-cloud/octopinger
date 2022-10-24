# Installation

## Development

### Cluster

```bash
minikube start -p octopinger --cpus=2 --disk-size=8gb --memory=4gb
```

```bash
minikube -p octopinger image load ghcr.io/ionos-cloud/octopinger/manager:v0.0.1
```