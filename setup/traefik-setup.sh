
helm repo add traefik https://traefik.github.io/charts

helm install traefik traefik/traefik -f traefik-values.yaml