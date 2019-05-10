.PHONY: build check-style go-dep go-build go-install fmt vet lint ensure-build-dir-exists docker-image install push-image

projectDir := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
buildDir := $(projectDir)/build
image := cassandra-webhook
pkgs := $(shell go list ./... | grep -v /vendor/ | grep -v /pkg/apis/ | grep -v /pkg/client/ | grep -v /test/e2e)

include $(projectDir)/../MakefileDocker

build: go-build go-install docker-image

install: build push-image

check-style: fmt vet lint

go-dep:
	@echo "== go-dep"
	dep ensure -v -vendor-only

go-build: ensure-build-dir-exists go-dep
	@echo "== build"
	go build -o $(buildDir)/bin/cassandra-webhook -v github.com/sky-uk/cassandra-operator/cassandra-webhook/cmd/webhook

go-install: go-build
	@echo "== install"
	cp -v $(buildDir)/bin/cassandra-webhook $(shell go env GOPATH)/bin/cassandra-webhook

fmt:
	go fmt ./...

vet: go-dep
	go vet $(pkgs)

lint:
	for pkg in $(pkgs); do \
		golint -set_exit_status $$pkg || exit 1; \
	done;

ensure-build-dir-exists:
	mkdir -p $(buildDir)

push-image: docker-image common-docker-push

docker-image: go-build common-docker-build
