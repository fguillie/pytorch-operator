# pytorch-operator

Kubernetes operator that manages NVIDIA PyTorch containers via the `PyTorchJob` CRD.
It creates and reconciles `Deployments` backed by NGC PyTorch images with GPU resource requests.

## Prerequisites

- Go >= 1.23
- Podman
- Helm 3
- kubectl with access to a Kubernetes cluster

---

## Build the image

Run from the **repository root** (`pytorch-operator/`):

> **Note:** `go.sum` must exist before building. If it is absent, generate it first:
> ```bash
> # from pytorch-operator/
> go mod tidy
> ```

Build for `linux/amd64` (pass `TARGETARCH` explicitly when building from an ARM host):

```bash
# from pytorch-operator/
podman build \
  --no-cache \
  --platform linux/amd64 \
  --build-arg TARGETARCH=amd64 \
  -t quay.io/fguillier/pytorch-operator:0.1.0 .
```

---

## Publish the image

Run from any directory:

```bash
podman push quay.io/fguillier/pytorch-operator:0.1.0
```

---

## Deploy via Helm

Run from the **repository root** (`pytorch-operator/`):

```bash
# from pytorch-operator/
helm install pytorch-operator ./helm \
  --namespace pytorch-operator --create-namespace \
  --set image.repository=quay.io/fguillier/pytorch-operator \
  --set image.tag=0.1.0 \
  --set image.pullPolicy=Always
```

Verify the operator is running:

```bash
kubectl get pods -n pytorch-operator
```

---

## PyTorchJob example

```yaml
apiVersion: pytorch.nvidia.com/v1alpha1
kind: PyTorchJob
metadata:
  name: pytorch-training
  namespace: default
spec:
  # NVIDIA NGC PyTorch container tag
  # Browse tags at https://catalog.ngc.nvidia.com/orgs/nvidia/containers/pytorch
  pytorchVersion: "24.01-py3"
  # nvidia.com/gpu resources per pod
  gpuCount: 2
  # Number of pods (default: 1)
  replicas: 1
  # Optional: override base image (default: nvcr.io/nvidia/pytorch)
  # image: "nvcr.io/nvidia/pytorch"
  # Optional: custom entrypoint
  # command: ["python", "-c"]
  # args: ["import torch; print(torch.cuda.device_count())"]
```

Apply from the **repository root** (`pytorch-operator/`):

```bash
# from pytorch-operator/
kubectl apply -f config/samples/pytorchjob_v1alpha1_example.yaml
```

Watch status:

```bash
kubectl get ptjob -A
```
