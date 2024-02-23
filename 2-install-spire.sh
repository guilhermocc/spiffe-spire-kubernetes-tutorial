kubectl config use-context kind-spire-example
helm upgrade --install -n spire-system spire-crds spire-crds --repo https://spiffe.github.io/helm-charts-hardened/ --create-namespace
helm upgrade --install -n spire-system spire spire --repo https://spiffe.github.io/helm-charts-hardened/ -f values.yaml
helm -n spire-system list
