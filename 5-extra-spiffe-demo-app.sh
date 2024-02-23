helm repo add spiffe-demo https://elinesterov.github.io/spiffe-demo-app

helm -n spiffe-demo install spiffe-demo  spiffe-demo/spiffe-demo-app --create-namespace -f .helm/spiffe-demo-app-values.yaml

kubectl -n spiffe-demo patch deployment spiffe-demo-app -p '{"spec":{"template":{"spec":{"containers":[{"name":"spiffe-demo-app","env":[{"name":"SPIFFE_ENDPOINT_SOCKET","value":"unix:///spiffe-workload-api/spire-agent.sock"}]}]}}}}'


# Wait for the deployment to be ready
while [ $(kubectl -n spiffe-demo get deployment spiffe-demo-app -o 'jsonpath={..status.conditions[?(@.type=="Available")].status}') != "True" ]; do
  echo "waiting for deployment"
  sleep 1
done

kubectl -n spiffe-demo port-forward deployment/spiffe-demo-app 8080:8080
