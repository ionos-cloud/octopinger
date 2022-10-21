# Installation

## Development

### Cluster

```bash
minikube start -p goldpinger --cpus=2 --disk-size=8gb --memory=4gb
```

```bash
minikube -p goldpinger image load ghcr.io/ionos-cloud/octopinger/manager:v0.0.1
```