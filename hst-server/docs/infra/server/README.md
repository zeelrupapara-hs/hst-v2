# hst-server on Kubernetes

Scaling notes and copy-paste manifests. Measured numbers come from a 10-core
laptop; re-measure on your nodes.

## Why StatefulSet, not Deployment

Each pod owns a shard of the session cache, and it learns which shard from its
ordinal. Deployments have no stable ordinal.

```
SHARD_ID    = pod ordinal, from the apps.kubernetes.io/pod-index label
SHARD_COUNT = replicas
```

Getting this wrong is not a correctness bug. Redis still answers every lookup,
you just waste memory and lose hit rate.

## Secret

```bash
kubectl create secret generic hst-auth \
  --from-literal=jwt-private-key="$(cd hst-server && make gen-keys | grep PRIVATE | cut -d= -f2-)" \
  --from-literal=password-pepper="$(openssl rand -base64 32)" \
  --from-literal=postgres-password='...'
```

The pepper is set once and never changed. Changing it invalidates every stored
password.

## StatefulSet

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: hst-server
spec:
  serviceName: hst-server          # needs the headless service below
  replicas: 3                      # must equal SHARD_COUNT
  selector:
    matchLabels: {app: hst-server}
  template:
    metadata:
      labels: {app: hst-server}
    spec:
      containers:
        - name: hst-server
          image: hst-server:1.0.0
          ports: [{containerPort: 8080}]
          env:
            - name: SHARD_ID       # 0, 1, 2 — the pod ordinal
              valueFrom:
                fieldRef: {fieldPath: metadata.labels['apps.kubernetes.io/pod-index']}
            - name: SHARD_COUNT
              value: "3"
            - name: MAX_ACCOUNT_PER_SHARD
              value: "100000"
            - name: AUTH_JWT_PRIVATE_KEY
              valueFrom:
                secretKeyRef: {name: hst-auth, key: jwt-private-key}
            - name: AUTH_PASSWORD_PEPPER
              valueFrom:
                secretKeyRef: {name: hst-auth, key: password-pepper}
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef: {name: hst-auth, key: postgres-password}
            - name: POSTGRES_SSL_MODE
              value: "require"          # verify-full if you pin the CA
            - name: REDIS_TLS
              value: "true"
            - name: SWAGGER_ENABLED
              value: "false"       # it exposes the whole api surface
            - name: CORS_ORIGINS
              value: "https://admin.example.com"
            - name: FIRST_MANAGER_PASSWORD
              value: ""            # empty after the first deploy
            - name: GOMAXPROCS     # sizes the argon2 semaphore
              valueFrom:
                resourceFieldRef: {resource: limits.cpu}
          resources:
            requests: {cpu: "2", memory: "1Gi"}
            limits:   {cpu: "2", memory: "2Gi"}
          livenessProbe:           # no database, so a slow query cannot kill the pod
            httpGet: {path: /api/v1/system/monitor/live, port: 8080}
            periodSeconds: 10
          readinessProbe:          # checks postgres and redis
            httpGet: {path: /api/v1/system/monitor/health, port: 8080}
            periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: hst-server
spec:
  clusterIP: None                  # headless, required by the StatefulSet
  selector: {app: hst-server}
  ports: [{port: 8080}]
```

## TLS

Terminate at the ingress and leave `HTTP_TLS_CERT` and `HTTP_TLS_KEY` empty.
The pod then serves plain http inside the cluster and the app still sends HSTS,
because it reads `X-Forwarded-Proto`.

To serve https from the pod instead, mount a certificate and set both:

```yaml
            - name: HTTP_TLS_CERT
              value: /tls/tls.crt
            - name: HTTP_TLS_KEY
              value: /tls/tls.key
          volumeMounts:
            - {name: tls, mountPath: /tls, readOnly: true}
      volumes:
        - name: tls
          secret: {secretName: hst-server-tls}
```

Setting only one of the pair is a boot error, so a half-configured pod fails
loudly instead of quietly serving http.

## Scaling is two commands, not one

```bash
kubectl scale statefulset hst-server --replicas=5
kubectl set env statefulset/hst-server SHARD_COUNT=5   # forgetting this wastes memory
```

No session is lost. Ownership reshuffles and hit rate dips while caches rewarm.

## Sizing

Memory per pod is the cache plus the hasher:

```
cache   = MAX_ACCOUNT_PER_SHARD x 552 bytes     100000 -> 53 MiB
argon2  = GOMAXPROCS x AUTH_ARGON2_MEMORY_KIB   2 x 64 MiB -> 128 MiB
```

`GOMAXPROCS` must come from the CPU limit as shown above. Left unset, Go reads
the host core count, so a 2-CPU pod on a 64-core node would allow 64 concurrent
hashes and ask for 4 GiB.

Throughput per pod, measured:

| path | rate | bound by |
|---|---|---|
| authenticated request | 5,000/s | nothing, zero SQL |
| refresh | ~110/s | postgres |
| login | ~82/s | argon2 cpu |

Login is deliberately slow. Size replicas for login rate, not request rate:
100k logins in 60s needs ~1,700/s, so ~21 pods of this class.

## HPA

Scale on CPU, since login is CPU bound.

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: hst-server
spec:
  scaleTargetRef: {apiVersion: apps/v1, kind: StatefulSet, name: hst-server}
  minReplicas: 3
  maxReplicas: 12
  metrics:
    - type: Resource
      resource: {name: cpu, target: {type: Utilization, averageUtilization: 70}}
```

The HPA does not update `SHARD_COUNT`. Set it to `maxReplicas` and accept a
lower hit rate, or reconcile it yourself.

## Postgres connections

Each pod opens `POSTGRES_MAX_CONN` (default 20).

```
20 pods x 20 = 400 connections
```

Postgres degrades past a few hundred. Put PgBouncer in transaction mode in
front before you get there.

## Redis

Sessions live here, so eviction means random logouts.

```
--maxmemory-policy noeviction
--appendonly yes
```

Every key is scoped by one id, so moving to Redis Cluster later needs no
application change.

## Rollout

The access token TTL is 2 hours and tokens carry a `kid`. To rotate the signing
key without logging anyone out: add the new key as verify-only, wait out the
access TTL, then promote it to signing.

Ordinary rollouts are safe at any time. Sessions live in Redis, so a pod
restart loses only its warm cache.
