IMAGE_PREFIX ?= ghcr.io/tetratelabs/
APP_NAME ?= teg-ext-auth-poc
TAG ?= latest

.PHONY: build-image push-image

build-image:
	docker buildx build . -t $(IMAGE_PREFIX)$(APP_NAME):$(TAG) --build-arg GO_LDFLAGS="$(GO_LDFLAGS)" --load

push-image:
	docker buildx build . -t $(IMAGE_PREFIX)$(APP_NAME):$(TAG) --build-arg GO_LDFLAGS="$(GO_LDFLAGS)" --load --push
