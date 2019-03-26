.DEFAULT_GOAL := travis
.PHONY: travis clean setup build install jar check dgoss docker-image push-image

TEST_REGISTRY ?= localhost:5000

projectDir := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
buildDir := $(projectDir)/build
gradle := $(projectDir)/../gradlew
gitRev := $(shell git rev-parse --short HEAD)

travis: clean install check

build: jar docker-image

install: build push-image

check: dgoss

setup:
	@echo "== setup"
ifeq (, $(shell which dgoss))
	@echo "dgoss not found, installing it"
	sudo curl -L https://github.com/aelsabbahy/goss/releases/download/v0.3.5/goss-linux-amd64 -o /usr/local/bin/goss \
        && sudo chmod +rx /usr/local/bin/goss \
        && sudo curl -L https://raw.githubusercontent.com/aelsabbahy/goss/v0.3.5/extras/dgoss/dgoss -o /usr/local/bin/dgoss \
        && sudo chmod +rx /usr/local/bin/dgoss
endif

jar:
	@echo "== jar"
	$(gradle) :cassandra-bootstrapper:exportStagedLib :cassandra-bootstrapper:bootstrapper:shadowJar

docker-image: jar
	@echo "== docker-image"
	docker build . -t $(TEST_REGISTRY)/cassandra-bootstrapper:v$(gitRev)

push-image: docker-image
	@echo "== push-image"
	docker push $(TEST_REGISTRY)/cassandra-bootstrapper:v$(gitRev)

dgoss:
	@echo "== dgoss"
	IMAGE_TO_TEST=$(TEST_REGISTRY)/cassandra-bootstrapper:v$(gitRev) ${projectDir}/test/runChecks.sh

clean:
	@echo "== clean"
	docker rm $(TEST_REGISTRY)/cassandra-bootstrapper:v$(gitRev) || true
	$(gradle) :cassandra-bootstrapper:clean :cassandra-bootstrapper:bootstrapper:clean :cassandra-bootstrapper:seed-provider:clean
