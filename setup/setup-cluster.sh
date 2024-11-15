#!/usr/bin/env bash

gcloud container \
  --project "knative-community" \
  clusters create "dave-cluster-1" \
  --region "us-east1" \
  --no-enable-basic-auth \
  --cluster-version "1.31.1-gke.1846000" \
  --release-channel "regular" \
  --machine-type "e2-standard-4" \
  --image-type "COS_CONTAINERD" \
  --disk-type "pd-balanced" \
  --disk-size "100" \
  --metadata disable-legacy-endpoints=true \
  --scopes "https://www.googleapis.com/auth/devstorage.read_only","https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/monitoring","https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management.readonly","https://www.googleapis.com/auth/trace.append" \
  --num-nodes "3" \
  --enable-ip-alias \
  --network "projects/knative-community/global/networks/default" \
  --subnetwork "projects/knative-community/regions/us-east1/subnetworks/default" \
  --no-enable-intra-node-visibility \
  --default-max-pods-per-node "110" \
  --enable-ip-access \
  --security-posture=disabled \
  --workload-vulnerability-scanning=disabled \
  --no-enable-master-authorized-networks \
  --no-enable-google-cloud-access \
  --addons HorizontalPodAutoscaling,HttpLoadBalancing \
  --enable-autoupgrade \
  --enable-autorepair \
  --max-surge-upgrade 1 \
  --max-unavailable-upgrade 0 \
  --binauthz-evaluation-mode=DISABLED \
  --no-enable-managed-prometheus \
  --gateway-api=standard




#   import (
# 	"fmt"
# 	"strings"
# 	"time"
# )

# func main() {
# 	for _, item := range strings.Split(strings.TrimSpace(data), "\n") {
# 		val := strings.TrimSpace(strings.Split(item, ":")[1])
# 		d, err := time.ParseDuration(strings.TrimSpace(val))
# 		if err != nil {
# 			panic(err)
# 		}

# 		fmt.Println(d.Seconds())
# 	}
# }
