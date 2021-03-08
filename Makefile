LOCAL_TAG = latest
DOCKER_FILE = Dockerfile

.PHONY: build
build:
	docker build -t $(LOCAL_TAG) $(DOCKER_FILE)

.PHONY: generate
generate:
	docker run --rm -v $(shell pwd):/local openapitools/openapi-generator-cli generate \
        -i https://raw.githubusercontent.com/All-Cups/highloadcup/main/goldrush/swagger.yaml \
        -g go \
        -o /local/openapi; \
    rm openapi/go.*