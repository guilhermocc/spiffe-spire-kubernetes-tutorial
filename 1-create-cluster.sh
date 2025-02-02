#!/bin/sh
set -o errexit

# create registry container unless it already exists
reg_name='kind-registry'
reg_port='8049'
running="$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)"
if [ "${running}" != 'true' ]; then
  docker run \
    -d --restart=always -p "${reg_port}:5000" --name "${reg_name}" \
    registry:2
fi

# Create the cluster.
# The containerdConfigPatches line is useful to configure access to the local registry.
# The ClusterConfiguration options are needed for nodes to access projected service
# account tokens.
# The example uses serviceAccountToken that is by default enabled in K8s 1.20
# To simplify the deployment, select a node image with K8s 1.20 or higher.
# The complete list of Kind images: https://github.com/kubernetes-sigs/kind/releases
cat <<EOF | kind create cluster --name spire-example -v 5 --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_name}:${reg_port}"]
nodes:
- role: control-plane
  image: kindest/node:v1.28.0
  kubeadmConfigPatches:
  - |
- role: worker
  image: kindest/node:v1.28.0
  kubeadmConfigPatches:
  - |
EOF



for node in $(kind get nodes); do
  kubectl annotate node "${node}" "kind.x-k8s.io/registry=localhost:${reg_port}";
done

# Install KubeShark for packet capture and analysis
kubectl config use-context kind-spire-example
helm repo add kubeshark https://helm.kubeshark.co
helm install kubeshark kubeshark/kubeshark

