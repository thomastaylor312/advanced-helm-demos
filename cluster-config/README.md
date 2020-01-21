# Cluster configuration using Charts
An example of configuring nodes in a Kubernetes cluster using a Chart. This is
just a simple example, but shows one possibility of using charts. Should you do
this in prod? Probably not (but never say never!), but it could give you an
interesting idea to build off of

## Prerequisites
You need to have a Kubernetes cluster available in which you have root access to
the nodes (meaning you can set a pod as privileged and with the ability to run
as root).

## Running the example
<!-- TODO: Make it change a DNS config and send a signal to the process to restart -->
