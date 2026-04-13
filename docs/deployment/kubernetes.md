# Kubernetes Deployment Guide

> Deploying OathMesh in a Kubernetes production environment.

## Overview

This guide covers deploying the OathMesh issuer and gateway in Kubernetes with proper security, scalability, and operational best practices.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Kubernetes Cluster                      │
│                                                                  │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐      │
│  │   Gateway   │     │   Issuer    │     │   Redis     │      │
│  │ Deployment  │ ──▶ │ Deployment │     │   Stateful  │      │
│  │  (2 replicas)│     │ (2 replicas)│     │   Set       │      │
│  └─────────────┘     └─────────────┘     └─────────────┘      │
│         │                   │                   │              │
│         └───────────────────┴───────────────────┘              │
│                         NetworkPolicy                           │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Kubernetes 1.24+
- Helm 3.x (optional, but recommended)
- Redis for replay cache backend

## Issuer Deployment

### Deployment Configuration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oathmesh-issuer
  namespace: oathmesh
spec:
  replicas: 2
  selector:
    matchLabels:
      app: oathmesh-issuer
  template:
    metadata:
      labels:
        app: oathmesh-issuer
    spec:
      containers:
      - name: issuer
        image: oathmesh/oathmesh:latest
        ports:
        - containerPort: 4000
        env:
        - name: OATHMESH_ISSUER
          value: "https://issuer.oathmesh.internal"
        - name: OATHMESH_PRIVATE_KEY
          valueFrom:
            secretKeyRef:
              name: oathmesh-private-key
              key: private-key
        - name: OATHMESH_TTL_DEFAULT
          value: "120"
        - name: OATHMESH_TTL_MAX
          value: "300"
        - name: OATHMESH_AUDIT_SINK
          value: "stdout"
        - name: REDIS_URL
          value: "redis://oathmesh-redis:6379/0"
        - name: OATHMESH_JWKS_CACHE_TTL
          value: "300"
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 256Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 4000
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /healthz
            port: 4000
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: oathmesh-issuer
  namespace: oathmesh
spec:
  selector:
    app: oathmesh-issuer
  ports:
  - port: 4000
    targetPort: 4000
  # Issuer should NOT be exposed publicly
  # Only accessible within the cluster
```

## Gateway Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oathmesh-gateway
  namespace: oathmesh
spec:
  replicas: 2
  selector:
    matchLabels:
      app: oathmesh-gateway
  template:
    metadata:
      labels:
        app: oathmesh-gateway
    spec:
      containers:
      - name: gateway
        image: oathmesh/oathmesh:latest
        command: ["./bin/oathmesh", "serve", "--gateway"]
        ports:
        - containerPort: 4000
        env:
        - name: OATHMESH_GATEWAY_UPSTREAM
          value: "http://your-upstream-service:8080"
        - name: OATHMESH_GATEWAY_AUDIENCE
          value: "https://api.internal"
        - name: OATHMESH_GATEWAY_ISSUERS
          value: "https://issuer.oathmesh.internal"
        - name: OATHMESH_GATEWAY_POLICY
          value: "/policy/production.json"
        - name: OATHMESH_JWKS_CACHE_TTL
          value: "300"
        volumeMounts:
        - name: policy
          mountPath: /policy
          readOnly: true
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 256Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 4000
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /healthz
            port: 4000
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: policy
        configMap:
          name: oathmesh-policy
```

## Redis for Replay Cache

```yaml
apiVersion: v1
kind: StatefulSet
metadata:
  name: oathmesh-redis
  namespace: oathmesh
spec:
  serviceName: oathmesh-redis
  replicas: 1
  selector:
    matchLabels:
      app: oathmesh-redis
  template:
    metadata:
      labels:
        app: oathmesh-redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
        volumeMounts:
        - name: data
          mountPath: /data
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            cpu: 200m
            memory: 128Mi
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: standard
      resources:
        requests:
          storage: 1Gi
```

## NetworkPolicy

Restrict access so only the gateway can reach the issuer:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: oathmesh-issuer-policy
  namespace: oathmesh
spec:
  podSelector:
    matchLabels:
      app: oathmesh-issuer
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: oathmesh-gateway
    ports:
    - protocol: TCP
      port: 4000
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: oathmesh-redis
    ports:
    - protocol: TCP
      port: 6379
  - to:
    - namespaceSelector: {}
      # Allow DNS
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
```

## Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: oathmesh-issuer-hpa
  namespace: oathmesh
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: oathmesh-issuer
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
    scaleUp:
      stabilizationWindowSeconds: 0
```

## Secrets Management

Create the private key secret:

```bash
kubectl create secret generic oathmesh-private-key \
  --from-file=private-key=/path/to/private.pem \
  -n oathmesh
```

Or use external secrets operator with your KMS (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault).

## Policy ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: oathmesh-policy
  namespace: oathmesh
data:
  production.json: |
    {
      "rules": [
        {
          "match": { "sub": "agent:*", "act": "read" },
          "allow": true
        },
        {
          "match": { "sub": "agent:*", "act": "write" },
          "allow": true
        },
        {
          "match": { "sub": "job:*" },
          "allow": true
        }
      ]
    }
```

## Deployment Checklist

- [ ] Create namespace: `kubectl create namespace oathmesh`
- [ ] Create private key secret
- [ ] Deploy Redis StatefulSet
- [ ] Deploy Issuer with 2 replicas
- [ ] Deploy Gateway with 2 replicas
- [ ] Apply NetworkPolicy
- [ ] Configure HPA
- [ ] Verify health endpoints: `kubectl get pods -n oathmesh`
- [ ] Test token minting: `./bin/oathmesh mint ...`
- [ ] Test gateway proxy: `curl http://gateway/inventory`

## Security Notes

1. **Never commit private keys** — Use Kubernetes Secrets or external KMS
2. **Issuer not publicly exposed** — Only gateway can reach it via NetworkPolicy
3. **Redis for replay cache** — Required for multi-instance deployments
4. **TLS required** — Issuer URL must be https:// in production
5. **Audit logging** — Set `OATHMESH_AUDIT_SINK=stdout` for cloud logging integration