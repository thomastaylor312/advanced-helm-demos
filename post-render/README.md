# Post Render Security Example
An example using post render to inject security requirements and side cars to
pods

## Running the example

```shell
$ cd kustomize
$ helm install security-test ../example-chart --post-renderer ./kustomize --debug --dry-run
```
