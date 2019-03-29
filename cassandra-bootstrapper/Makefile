.DEFAULT_GOAL := travis
.PHONY: travis clean setup build install jar check dgoss docker-image push-image release

projectDir := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
gradle := $(projectDir)/../gradlew
image := cassandra-bootstrapper

include $(projectDir)/../MakefileDocker

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

docker-image: jar common-docker-build

push-image: docker-image common-docker-push

dgoss:
	@echo "== dgoss"
	IMAGE_TO_TEST=$(dockerTestImage) ${projectDir}/test/runChecks.sh

clean: common-docker-clean
	@echo "== clean"
	$(gradle) :cassandra-bootstrapper:clean :cassandra-bootstrapper:bootstrapper:clean :cassandra-bootstrapper:seed-provider:clean

release: common-docker-release
