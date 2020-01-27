module github.com/thomastaylor312/advanced-helm-demos/controller

go 1.13

require (
	helm.sh/helm/v3 v3.0.2
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
