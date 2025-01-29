#!/bin/bash

# Patch greeter-client deployment
kubectl patch deployment greeter-client -n workload --type='json' -p='[
  {
    "op": "add",
    "path": "/spec/template/spec/containers/0/volumeMounts",
    "value": [
      {
        "name": "spiffe-workload-api",
        "mountPath": "/spiffe-workload-api",
        "readOnly": true
      }
    ]
  },
  {
    "op": "add",
    "path": "/spec/template/spec/volumes",
    "value": [
      {
        "name": "spiffe-workload-api",
        "csi": {
          "driver": "csi.spiffe.io",
          "readOnly": true
        }
      }
    ]
  },
  {
    "op": "add",
    "path": "/spec/template/spec/containers/0/env/-",
    "value": {
      "name": "SPIFFE_ENDPOINT_SOCKET",
      "value": "unix:///spiffe-workload-api/spire-agent.sock"
    }
  },
  {
    "op": "add",
    "path": "/spec/template/spec/containers/0/env/-",
    "value": {
      "name": "AUTHORIZED_SPIFFE_IDS",
      "value": "spiffe://example.org/ns/workload/sa/greeter-server-sa"
    }
  }
]'

# Patch greeter-server deployment
kubectl patch deployment greeter-server -n workload --type='json' -p='[
  {
    "op": "add",
    "path": "/spec/template/spec/containers/0/volumeMounts",
    "value": [
      {
        "name": "spiffe-workload-api",
        "mountPath": "/spiffe-workload-api",
        "readOnly": true
      }
    ]
  },
  {
    "op": "add",
    "path": "/spec/template/spec/volumes",
    "value": [
      {
        "name": "spiffe-workload-api",
        "csi": {
          "driver": "csi.spiffe.io",
          "readOnly": true
        }
      }
    ]
  },
  {
    "op": "add",
    "path": "/spec/template/spec/containers/0/env/-",
    "value": {
      "name": "SPIFFE_ENDPOINT_SOCKET",
      "value": "unix:///spiffe-workload-api/spire-agent.sock"
    }
  }
]'
