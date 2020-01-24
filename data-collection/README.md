# Data Collection using Helm
This is an example of how to use Helm to collect data about all your releases and the status of the objects they created and output it as a JSON object. It is fairly bare bones, but would serve as a good scaffold for actually doing something useful in production. You then have the flexibility to output it in any format for your data collection needs

## Prerequisites
You need to have a running Kubernetes cluster with a valid kubeconfig pointing to it. That cluster should have several Helm releases running on it, and for best demo-ability, some of the objects it created should not be in a ready state.

Because this is a Go example, you will also need a working Go environment. This was built and tested with Go 1.13

## Running the example

```shell
$ go run main.go                                     
{
    "KindCount": {
        "/v1/ConfigMap": 6,
        "/v1/PersistentVolumeClaim": 1,
        "/v1/Secret": 3,
        "/v1/Service": 4,
        "apps/v1/DaemonSet": 1,
        "apps/v1/Deployment": 1,
        "apps/v1/StatefulSet": 3
    },
    "TotalReleases": 3,
    "NumReady": 4,
    "NumNotReady": 1
}
```
