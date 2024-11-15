#!/usr/bin/env bash

kubectl apply -f https://projectcontour.io/quickstart/contour-gateway-provisioner.yaml

kubectl apply -f - <<EOF
kind: GatewayClass
apiVersion: gateway.networking.k8s.io/v1
metadata:
  name: contour
spec:
  controllerName: projectcontour.io/gateway-controller
EOF

kubectl apply -n default -f - <<EOF
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1
metadata:
  name: gateway
spec:
  gatewayClassName: contour
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
EOF