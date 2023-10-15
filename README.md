# Canary Bot

HTTP-based (gRPC) connectivity monitoring from node to node.

Run one Canary Bot on each distributed host to create a mesh.

Each bot will gather information about the connectivity to each other.

Current measurement samples:

- Round-trip-time with TCP, TLS handshake and request
- Round-trip-time TCP request

Every bot exposes an API (REST and RPC) for consuming measurement samples.

Each bot will provide samples from every node.

# Get your Canary

Get the latest release from the [release page](https://gitlab.devops.telekom.de/caas/canary-bot/-/releases).

## Binary

Download the binary from GitLab.
For authorization use e.g. a [personal token](https://docs.gitlab.com/ee/api/#authentication).

````
curl -c https://gitlab.devops.telekom.de/api/v4/projects/124625/packages/generic/cbot/${VERSION}/cbot
````

## Image

Get the latest image from the MTR.

````
docker image pull mtr.devops.telekom.de/caas/canary-bot:latest
docker image pull mtr.devops.telekom.de/caas/canary-bot:${VERSION}
````

# Usage

Run `cbot --help` for further information.

## The Network

To separate different scenarios like starting the bot on a dedicated host or running it on a Kubernetes cluster we
introduced the `join-address` and `listen-address` flag.

JoinMesh Request to tell the joining mesh 'who I am' - the public connection point:
`join-address` (optional; eg. test.de:443, localhost:8080) > external IP (form network interface)

Listen server address & port - the listening settings for the grpc server:
`listen-address` (optional; eg. 10.34.0.10, localhost) > external IP (form network interface)

### 1. Scenario: Kubernetes cluster

[edge-terminating TLS - 2 targets - different join & listen address for Kubernetes scenario]

The kubernetes ingress controller will listen on bird-owl.com on port 443 (https).
It redirects incoming http request via a service to the running pod listening to localhost on port 8080.

```
cbot --name owl --join-address bird-owl.com:443 --listen-adress localhost --listen-port 8080 --api-port 8081 -t bird-goose.com:443 -t bird-eagle.net:8080 --ca-cert-path path/to/cert.cer
```

### 2. Scenario: Dedicated host

[mutual TLS - 2 targets - join & listen-address is external IP from network interface]

The bot is running on a public ip (x.x.x.x) and listens on port 8081 for mesh requests.

```
cbot --name swan -t bird-goose.com:443 -t bird-eagle.net:8080 --ca-cert-path path/to/cert.cer --server-cert-path path/to/cert.cer --server-key ZWFzdGVyZWdn
```

## TLS

1. No TLS

- nothing todo

2. edge terminated TLS

E.g. in a Kubenetes Cluster with NGINX Ingress Controller

- Client: needs CA Cert
- Server: nothing todo, TLS is terminated before reaching server
- use: `ca-cert` flag

3. e2e Mutual TLS

- Client: needs CA Cert
- Server: needs Server Cert & Server Key
- use: `ca-cert`, `server-cert`, `server-key` flags

# Mesh

![the mesh](mesh.drawio.png)

# Logic

(comming soon [here](logic.md))

# Dev

(comming soon [here](dev.md))

