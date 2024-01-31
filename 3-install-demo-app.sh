cd greeter
make docker-build
make deploy

kubectl config use-context kind-spire-example

kubectl create namespace workload
kubectl apply -f server-deploy.yaml
kubectl apply -f client-deploy.yaml
