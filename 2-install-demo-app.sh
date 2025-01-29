make deploy

kubectl config use-context kind-spire-example

kubectl create namespace workload --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f greeter/server-deploy.yaml
kubectl apply -f greeter/client-deploy.yaml
