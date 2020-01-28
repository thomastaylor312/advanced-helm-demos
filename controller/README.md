# Helm Controller
A simple controller and CRD example using Helm. Once again, not production
quality, but a good starting point for anyone wanting to write their own

## Prerequisites
If you are just running the example, you'll need:
- GNU `make`
- `helm`
- A running Kubernetes cluster

If you are wanting to edit and build things yourself, you'll also need:
- Go 1.13
- Docker

## Running the example
You can run the latest example I've built into a Docker container and pushed up
to Docker Hub. It should be as simple as running:

```shell
$ make deploy
```

Once the controller is running, you can apply the various Helm CRDs in the
`examples/` directory using `kubectl apply -f`. Once you create those objects,
you'll see the various services, pods, and deployments spin up. You should see
the same number of pods as specified in the `replicas` field of the CRD.

## Building your own
If you want to build your own image or mess around with the code, you need to
log in to a container registry of your choice and set the `IMAGE_ORG`,
`IMAGE_NAME`, and `IMAGE_VERSION` (optional) environment variables. Then run:

```shell
$ make clean push deploy
```

If you have edited the Go API types, make sure to run `make codegen` before you
build
