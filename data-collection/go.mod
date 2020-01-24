module github.com/thomastaylor312/advanced-helm-demos/data-collection

go 1.13

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

require (
	helm.sh/helm/v3 v3.0.2
	k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
)
