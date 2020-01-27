module github.com/thomastaylor312/advanced-helm-demos/controller

go 1.13

require (
	helm.sh/helm/v3 v3.0.2
	k8s.io/apimachinery v0.17.2
	k8s.io/code-generator v0.17.2
	k8s.io/sample-controller v0.17.2 // indirect
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
