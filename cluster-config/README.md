# Cluster configuration using Charts
An example of configuring nodes in a Kubernetes cluster using a Chart. This is
just a simple example, but shows one possibility of using charts. Should you do
this in prod? Probably not (but never say never!), but it could give you an
interesting idea to build off of

## Prerequisites
You need to have a Kubernetes cluster available in which you have root access to
the nodes (meaning you can set a pod as privileged and with the ability to run
as root). This example only works on Linux nodes

## Running the example
To run this example, run the following command (assuming you are in this
directory)

```shell
$ helm install configure -n kube-system ./
```

You can then check the logs of the pod to see it running and SSH into any of
your nodes to see the new "hello-world.service" logs
