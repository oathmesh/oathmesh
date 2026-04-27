# Envoy `ext_authz` Integration

OathMesh provides a standalone binary (`oathmesh-envoy`) that implements the `envoy.service.auth.v3.Authorization` gRPC interface. This allows Envoy to delegate authentication and authorization to OathMesh via the `ext_authz` filter.

## Architecture

The `oathmesh-envoy` service runs alongside your Envoy proxy (e.g., as a sidecar container in Kubernetes). Envoy intercepts incoming requests and forwards the `Authorization` header to the `oathmesh-envoy` service.

1. **Token Extraction:** The service extracts the token (Bearer or OathMesh prefix).
2. **Zero-Trust Verification:** It validates the cryptographic signature, audience, issuers, expiration, and checks the token against replay and revocation caches.
3. **Context Injection:** Upon successful verification, the service instructs Envoy to inject the `X-OathMesh-Subject`, `X-OathMesh-Action`, and `X-OathMesh-Issuer` headers into the upstream request.

## Running the Service

Deploy the `oathmesh-envoy` binary. It exposes a gRPC service on port `50051` by default.

```bash
ENVOY_AUTHZ_PORT=50051 \
OATHMESH_ISSUER=https://issuer.local \
OATHMESH_GATEWAY_AUDIENCE=https://my-service.local \
OATHMESH_GATEWAY_ISSUERS=https://issuer.local \
./oathmesh-envoy
```

## Envoy Configuration

Configure Envoy to use the `envoy.filters.http.ext_authz` filter.

```yaml
static_resources:
  listeners:
  - address:
      socket_address:
        address: 0.0.0.0
        port_value: 8080
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          route_config: ...
          http_filters:
          - name: envoy.filters.http.ext_authz
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz
              grpc_service:
                envoy_grpc:
                  cluster_name: oathmesh_ext_authz
                timeout: 0.5s
              include_peer_certificate: true
              clear_route_cache: true
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router

  clusters:
  - name: oathmesh_ext_authz
    connect_timeout: 0.25s
    type: STRICT_DNS
    typed_extension_protocol_options:
      envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
        "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
        explicit_http_config:
          http2_protocol_options: {}
    load_assignment:
      cluster_name: oathmesh_ext_authz
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1 # Point to the oathmesh-envoy service
                port_value: 50051
```

Once configured, any request missing a valid OathMesh token will be rejected by Envoy with a `401 Unauthorized` or `403 Forbidden` response. Valid requests will reach your upstream service enriched with the `X-OathMesh-*` headers.
