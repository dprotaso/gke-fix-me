#!/usr/bin/env bash


gcloud container clusters update dave-cluster-1 \
    --project "knative-community" \
    --location=us-east1 \
    --gateway-api=standard

kubectl apply -f - <<EOF
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1
metadata:
  name: gateway
spec:
  gatewayClassName: gke-l7-regional-external-managed
  listeners:
  - name: http
    protocol: HTTP
    port: 80
EOF

# OR

kubectl apply -f - <<EOF
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1
metadata:
  name: gateway
spec:
  gatewayClassName: gke-l7-gxlb
  listeners:
  - name: http
    protocol: HTTP
    port: 80
EOF