IMAGE_BUILD_CMD ?= "docker"
IMAGE_REGISTRY ?= "quay.io"
REGISTRY_NAMESPACE ?= "alukiano"
IMAGE_TAG ?= "4.8-snapshot"

POD_RESTART_CONTROLLER_IMAGE ?= "$(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/pod-restarter:$(IMAGE_TAG)"

.PHONY: build
build:
	./hack/build.sh

.PHONY: deps-update
deps-update:
	go mod tidy && go mod vendor

.PHONY: build-image
build-image: build
	$(IMAGE_BUILD_CMD) build --no-cache -f Dockerfile -t $(POD_RESTART_CONTROLLER_IMAGE) --build-arg BIN_DIR="_output/bin" .

.PHONY: push-image
push-image: build-image
	$(IMAGE_BUILD_CMD) push $(POD_RESTART_CONTROLLER_IMAGE)

.PHONY: clean
clean:
	rm -rf _output
