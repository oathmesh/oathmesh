# Kong API Gateway Integration

OathMesh integrates natively with Kong API Gateway utilizing the **Kong Go PDK**. This allows the 14-step verification pipeline to execute as a high-performance external process rather than relying on a Lua rewrite.

## Architecture & Topology

Unlike traditional Lua plugins which run in-process within NGINX/Kong, Kong Go plugins run as an external process. Kong communicates with the `oathmesh-kong` plugin over `msgpack` via a Unix socket or a local TCP port.

```text
Incoming Request -> Kong API Gateway -> (msgpack) -> OathMesh Kong Go Plugin
                                                        |
                 <- (X-OathMesh-* Headers) <------------+
Upstream Service <-
```

## Compilation

Build the plugin binary. The binary will act as the plugin server.

```bash
cd plugins/kong
go build -o oathmesh-plugin main.go
```

## Deployment

In your Kong environment (e.g., Kubernetes or Docker Compose), you must deploy the built binary alongside the Kong process.

1. Mount the `oathmesh-plugin` binary into your Kong container (or build a custom Kong image containing it).
2. Configure the Kong environment variables to point to the external plugin socket/executable.

```yaml
environment:
  KONG_PLUGINS: bundled,oathmesh-plugin
  KONG_PLUGINSERVER_NAMES: oathmesh-plugin
  KONG_PLUGINSERVER_OATHMESH_PLUGIN_SOCKET: /usr/local/kong/oathmesh.socket
  KONG_PLUGINSERVER_OATHMESH_PLUGIN_START_CMD: /path/to/oathmesh-plugin
  KONG_PLUGINSERVER_OATHMESH_PLUGIN_QUERY_CMD: /path/to/oathmesh-plugin -dump
```

*Note: For highly available setups in Kubernetes, the plugin server can run as a sidecar container sharing the same network namespace and communicating over a shared volume for the Unix socket or via localhost TCP.*

## Configuration

Once the plugin is registered, you can enable it on a Service, Route, or Global level via the Kong Admin API or decK.

### Declarative Configuration (decK)

```yaml
_format_version: "3.0"
services:
- name: my-upstream-service
  url: http://my-upstream:8080
  plugins:
  - name: oathmesh-plugin
    config:
      audience: "https://my-service.local"
      issuers: "https://issuer.local"
```

## Result

Kong will intercept the `Authorization` header, pass it to the OathMesh plugin, and upon successful validation, inject the `X-OathMesh-Subject`, `X-OathMesh-Action`, and `X-OathMesh-Issuer` headers before forwarding the request to the upstream service.
