LOCAL_TAG = gold-rush:latest
REMOTE_TAG = stor.highloadcup.ru/rally/electric_albatross

.PHONY: build
build:
	docker build -t $(LOCAL_TAG)
    docker tag $(LOCAL_TAG) $(REMOTE_TAG)
    docker push $(REMOTE_TAG)

.PHONY: generate
generate:
	docker run --rm -v $(shell pwd):/local openapitools/openapi-generator-cli generate \
        -i https://raw.githubusercontent.com/All-Cups/highloadcup/main/goldrush/swagger.yaml \
        -g go \
        -o /local/openapi; \
    rm openapi/go.*