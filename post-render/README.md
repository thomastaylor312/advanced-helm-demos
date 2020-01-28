# Post Render Security Example
An example using post render to inject security requirements and side cars to
pods

## Prerequisites
- `kustomize` 3.5+ installed in your `PATH`
- A running Kubernetes cluster (or use `helm template` instead of `helm install` in the command below)
- Helm 3.1

## Running the example

```shell
$ cd kustomize
$ helm install security-test ../example-chart --post-renderer ./kustomize --debug --dry-run
```
