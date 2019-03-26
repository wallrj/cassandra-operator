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
	$(gradle) :fake-cassandra-docker:shadowJar

docker-image: jar
	@echo "== docker-image"
	docker build . -t $(TEST_REGISTRY)/fake-cassandra:v$(gitRev)

push-image: docker-image
	@echo "== push-image"
	docker push $(TEST_REGISTRY)/fake-cassandra:v$(gitRev)

dgoss:
	@echo "== dgoss"
	dgoss run $(TEST_REGISTRY)/fake-cassandra:v$(gitRev)

clean:
	@echo "== clean"
	docker rm $(TEST_REGISTRY)/fake-cassandra:v$(gitRev) || true
	$(gradle) :fake-cassandra-docker:clean
