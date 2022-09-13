# Usage

Run `cbot --help` for futher information.

## the network

To separate different szenarios like starting the bot on a dedicated host or e.g. a Kubernetes cluster we introduced the `join-address` and `listen-address` flag.

JoinMesh Request to tell the joining mesh who I am - the public connection point:
`join-address` (optional; eg. test.de:443, localhost:8080) > external IP (form network interface)

Listen server address & port - the real listening settings of the grpc server:
`listen-address` (optional; eg. 10.34.0.10, localhost) > external IP (form network interface)

### 1. Szenario: Kubernetes cluster

The kubernetes ingress controller will listen on my-ingress-domain.com on port 443 (https). It redirects incomming http request via a service to the running pod listening on its internal ip on port 8080.

```
... --join-address my-ingress-domain.com:443 --listen-port 8080 ...
```

### 2. Szenario: Dedicated host

The bot is running on a public ip (x.x.x.x) and listens on port 8081 for mesh requests.

```
... --listen-port 8081 ...
```

## TLS


1. No TLS

- nothing todo

2. edge terminated TLS

- eg. in a Kubenetes Cluster with NGINX Ingress Controller
- Client: needs CA Cert
- Server: nothing todo, TLS is terminated before reaching server
- use: `ca-cert` flag

2. e2e Mutal TLS

- Client: needs CA Cert
- Server: needs Server Cert & Server Key
- use: `ca-cert`, `server-cert`, `server-key` flags

# Logic

Have a deeper look at the canary logic [here](logic.md)

