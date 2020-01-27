.PHONY: codegen build-image push deploy

IMAGE_ORG ?= thomastaylor312
IMAGE_NAME ?= demo-helm-controller
IMAGE_VERSION ?= $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null || git rev-parse --short HEAD)

codegen:
	./hack/update-codegen.sh

build: bin/controller-darwin bin/controller-windows bin/controller-linux

bin/controller-darwin:
	GOOS=darwin go build -o bin/controller-darwin ./*.go

bin/controller-windows:
	GOOS=windows go build -o bin/controller-windows ./*.go

bin/controller-linux:
	GOOS=linux go build -o bin/controller-linux ./*.go
	
build-image: build
	docker build -t $(IMAGE_ORG)/$(IMAGE_NAME):$(IMAGE_VERSION) .

push: build-image
	docker push $(IMAGE_ORG)/$(IMAGE_NAME):$(IMAGE_VERSION)

deploy:
	helm upgrade --install demo-controller ./charts/controller --set image.name=$(IMAGE_ORG)/$(IMAGE_NAME):$(IMAGE_VERSION),serviceAccount.create=true

clean:
	rm -rf bin/
