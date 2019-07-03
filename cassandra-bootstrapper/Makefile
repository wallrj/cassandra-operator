.DEFAULT_GOAL := all
.PHONY: all clean build install jar check dgoss docker-image push-image release check-style

projectDir := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
gradle := $(projectDir)/../gradlew
image := cassandra-bootstrapper

include $(projectDir)/../MakefileDocker

all: clean install check

build: check-style jar docker-image

install: build push-image

check: dgoss

jar:
	@echo "== jar"
	$(gradle) :cassandra-bootstrapper:exportStagedLib :cassandra-bootstrapper:bootstrapper:shadowJar

docker-image: jar common-docker-build

push-image: docker-image common-docker-push

dgoss:
	@echo "== dgoss"
	IMAGE_TO_TEST=$(dockerTestImage) ${projectDir}/test/runChecks.sh

clean: common-docker-clean
	@echo "== clean"
	$(gradle) :cassandra-bootstrapper:clean :cassandra-bootstrapper:bootstrapper:clean :cassandra-bootstrapper:seed-provider:clean

release: common-docker-release

check-style:
	@echo "== check-style not implemented"
