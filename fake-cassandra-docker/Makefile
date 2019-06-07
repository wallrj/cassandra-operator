.DEFAULT_GOAL := all
.PHONY: all clean setup build install jar check dgoss docker-image push-image release check-style

projectDir := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
gradle := $(projectDir)/../gradlew
image := fake-cassandra

include $(projectDir)/../MakefileDocker

all: clean install check

build: check-style jar docker-image

install: build push-image

check-style:
	@echo "== check-style not implemented"

check: dgoss

setup:
	@echo "== setup"

jar:
	@echo "== jar"
	$(gradle) :fake-cassandra-docker:shadowJar

docker-image: jar common-docker-build

push-image: docker-image common-docker-push

dgoss:
	@echo "== dgoss"
	dgoss run --env "NODE_LISTEN_ADDRESS=192.0.2.100" $(dockerTestImage)

clean: common-docker-clean
	@echo "== clean"
	$(gradle) :fake-cassandra-docker:clean

release: common-docker-release
