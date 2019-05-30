.DEFAULT_GOAL := all
.PHONY: all check-style clean-all build-all setup setup-all check-style-all run-local-registry

TEST_REGISTRY ?= localhost:5000
POD_START_TIMEOUT ?= 120s
KUBE_CONTEXT ?= dind
USE_MOCK ?= true
FAKE_CASSANDRA_IMAGE ?= $(TEST_REGISTRY)/fake-cassandra:v$(gitRev)
CASSANDRA_BOOTSTRAPPER_IMAGE ?= $(TEST_REGISTRY)/cassandra-bootstrapper:v$(gitRev)
CASSANDRA_SNAPSHOT_IMAGE ?= $(TEST_REGISTRY)/cassandra-snapshot:v$(gitRev)
NAMESPACE ?= test-cassandra-operator
GINKGO_NODES ?= 0
GINKGO_COMPILERS ?= 0

gitRev := $(shell git rev-parse --short HEAD)
projectDir := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))

all: setup install dind check

setup: check-system-dependencies setup-all run-local-registry

build: build-all

install: install-all

check: check-all

clean: clean-all

release: release-all

check-style: check-style-all

run-local-registry:
	docker run -d --name=dind-registry --rm -p 5000:5000 registry:2

check-system-dependencies:
	@echo "== check-system-dependencies"
ifeq (, $(shell which go))
	$(error "golang not found in PATH")
endif
ifeq (, $(shell which rsync))
	$(error "rsync not found in PATH")
endif
ifeq (, $(shell which docker))
	$(error "docker not found in PATH")
endif
ifeq (, $(shell which dgoss))
	$(error "dgoss not found in PATH")
endif
ifeq (, $(shell which java))
	$(error "java not found in PATH")
endif
ifeq (, $(shell which kubectl))
	$(error "kubectl not found in PATH")
endif

check-all:
	@echo "== check-all"
	$(MAKE) -C cassandra-bootstrapper check
	$(MAKE) -C fake-cassandra-docker check
	GINKGO_NODES=$(GINKGO_NODES) GINKGO_COMPILERS=$(GINKGO_COMPILERS) KUBE_CONTEXT=$(KUBE_CONTEXT) TEST_REGISTRY=$(TEST_REGISTRY) FAKE_CASSANDRA_IMAGE=$(FAKE_CASSANDRA_IMAGE) USE_MOCK=$(USE_MOCK) $(MAKE) -C cassandra-snapshot check
	GINKGO_NODES=$(GINKGO_NODES) GINKGO_COMPILERS=$(GINKGO_COMPILERS) KUBE_CONTEXT=$(KUBE_CONTEXT) TEST_REGISTRY=$(TEST_REGISTRY) FAKE_CASSANDRA_IMAGE=$(FAKE_CASSANDRA_IMAGE) CASSANDRA_BOOTSTRAPPER_IMAGE=$(CASSANDRA_BOOTSTRAPPER_IMAGE) CASSANDRA_SNAPSHOT_IMAGE=$(CASSANDRA_SNAPSHOT_IMAGE) USE_MOCK=$(USE_MOCK) POD_START_TIMEOUT=$(POD_START_TIMEOUT) $(MAKE) -C cassandra-operator check

build-all:
	@echo "== build-all"
	$(MAKE) -C fake-cassandra-docker build
	$(MAKE) -C cassandra-bootstrapper build
	$(MAKE) -C cassandra-snapshot build
	$(MAKE) -C cassandra-operator build

install-all:
	@echo "== install-all"
	$(MAKE) -C fake-cassandra-docker install
	$(MAKE) -C cassandra-bootstrapper install
	$(MAKE) -C cassandra-snapshot install
	$(MAKE) -C cassandra-operator install

clean-all:
	@echo "== clean-all"
	$(MAKE) -C fake-cassandra-docker clean
	$(MAKE) -C cassandra-bootstrapper clean
	$(MAKE) -C cassandra-snapshot clean
	$(MAKE) -C cassandra-operator clean

setup-all:
	@echo "== setup-all"
	$(MAKE) -C fake-cassandra-docker setup
	$(MAKE) -C cassandra-bootstrapper setup
	$(MAKE) -C cassandra-snapshot setup
	$(MAKE) -C cassandra-operator setup

dind:
	@echo "== recreate dind cluster"
	NAMESPACE=$(NAMESPACE) $(projectDir)/test-kubernetes-cluster/recreate-dind-cluster.sh

release-all:
	@echo "== release-all"
	$(MAKE) -C fake-cassandra-docker release
	$(MAKE) -C cassandra-bootstrapper release
	$(MAKE) -C cassandra-snapshot release
	$(MAKE) -C cassandra-operator release

check-style-all:
	@echo "== check-style-all"
	$(MAKE) -C fake-cassandra-docker check-style
	$(MAKE) -C cassandra-bootstrapper check-style
	$(MAKE) -C cassandra-snapshot check-style
	$(MAKE) -C cassandra-operator check-style
