.DEFAULT_GOAL := travis
.PHONY: travis clean setup build install jar check dgoss docker-image push-image release

projectDir := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
gradle := $(projectDir)/../gradlew
image := fake-cassandra

include $(projectDir)/../MakefileDocker

travis: clean install check

build: jar docker-image

install: build push-image

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
	dgoss run $(dockerTestImage)

clean: common-docker-clean
	@echo "== clean"
	$(gradle) :fake-cassandra-docker:clean

release: common-docker-release
