#!/bin/bash

kubectl delete -k kubernetes/overlays/development/
cd ../backend || exit 1

for dirname in *_service gateway; do
    echo "Processing $dirname"
    kube_name=${dirname//_/-}
    docker build -f "$dirname"/Dockerfile -t qf-"$kube_name":latest .
    minikube image rm qf-"$kube_name"
    minikube image load qf-"$kube_name" || /dev/null
   #kubectl rollout restart deployment/"$kube_name"-deployment
done

echo "Processing smart-balancer"
cd ../deploy/smart-balancer || exit 1
docker build -f deploy/Dockerfile -t smart-balancer:latest .
minikube image load smart-balancer
#kubectl rollout restart deployment/smart-balancer

cd ../ || exit 1
kubectl apply -k kubernetes/overlays/development/