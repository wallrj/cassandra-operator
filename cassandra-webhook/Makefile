.DEFAULT_GOAL := all
.PHONY: all clean install check build fmt vet lint validate-licences test e2e-te setup go-build go-install docker-image push-image release check-style \
	validate-crds generate-crds \
	validate-types generate-types

# Default context is running locally against a dind cluster
KUBE_CONTEXT ?= dind
POD_START_TIMEOUT ?= 120s
USE_MOCK ?= true
FAKE_CASSANDRA_IMAGE ?= $(TEST_REGISTRY)/fake-cassandra:v$(gitRev)
CASSANDRA_BOOTSTRAPPER_IMAGE ?= $(TEST_REGISTRY)/cassandra-bootstrapper:v$(gitRev)
CASSANDRA_SIDECAR_IMAGE ?= $(TEST_REGISTRY)/cassandra-sidecar:v$(gitRev)
CASSANDRA_SNAPSHOT_IMAGE ?= $(TEST_REGISTRY)/cassandra-snapshot:v$(gitRev)
DOMAIN ?= localhost
IMAGE_TO_DEPLOY ?= $(dockerTestImage)
NAMESPACE ?= test-cassandra-operator
GINKGO_NODES ?= 0
GINKGO_COMPILERS ?= 0

projectDir := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
buildDir := $(projectDir)/build
image := cassandra-webhook
junitReportDir := $(buildDir)/junit-reports
pkgs := $(shell go list ./... | grep -v /vendor/ | grep -v /pkg/apis/ | grep -v /pkg/client/ | grep -v /test/e2e)
# testPkgs := $(shell go list ./... | grep /test/e2e)

include $(projectDir)/../MakefileDocker

all: clean install check

build: go-build check-style validate-licences validate-crds validate-types docker-image

install: build go-install push-image

check-style: fmt vet lint

check: test deploy-operator integration-test e2e-test

ensure-go-modules:
ifeq ('$(shell go env | grep -o "GOMOD.*")', 'GOMOD=""')
	$(error "go modules must be enabled: export GO111MODULE=on, or run from outside $$GOPATH")
else
	@echo "go modules enabled"
endif

setup: ensure-go-modules
	@echo "== setup"
	go install \
		github.com/sky-uk/licence-compliance-checker \
		github.com/onsi/ginkgo/ginkgo \
		golang.org/x/tools/cmd/goimports \
		golang.org/x/lint/golint \
		sigs.k8s.io/controller-tools/cmd/controller-gen

go-build: ensure-build-dir-exists ensure-go-modules
	@echo "== build"
	go build -o $(buildDir)/bin/cassandra-webhook -v ./cmd/main.go
	# KUBE_CONTEXT=ignore go test -run xxxxx $(testPkgs) # build the test code but don't run any tests yet

go-install: go-build
	@echo "== install"
	cp -v $(buildDir)/bin/cassandra-webhook $(shell go env GOPATH)/bin/cassandra-webhook

fmt:
	go fmt ./...

vet:
	go vet $(pkgs)

lint:
	for pkg in $(pkgs); do \
		golint -set_exit_status $$pkg || exit 1; \
	done;

ensure-build-dir-exists:
	mkdir -p $(buildDir)

ensure-test-report-dir-exists: ensure-build-dir-exists
	mkdir -p $(junitReportDir)

test: ensure-test-report-dir-exists
	@echo "== test"
	ginkgo -p -nodes=$(GINKGO_NODES) -r -compilers=$(GINKGO_COMPILERS) --v --progress pkg cmd -- -junit-report-dir $(junitReportDir)

docker-image: go-build common-docker-build

push-image: docker-image common-docker-push

validate-licences:
	@echo "== validate licences"
	set -e ;\
	restricted=$$(paste -s -d ',' restricted-licences.txt) ;\
	licence-compliance-checker -L error -E -r $$restricted \
		-m github.com/sky-uk/cassandra-operator/cassandra-webhook=BSD-3-Clause \
		-m github.com/sky-uk/cassandra-operator/cassandra-operator=BSD-3-Clause \
		-m sigs.k8s.io/yaml=BSD-3-Clause \
		-m github.com/hashicorp/golang-lru=MPL-2.0 \
		-m github.com/gogo/protobuf=BSD-3-Clause \
		--check-go-modules ;\

clean: common-docker-clean
	@echo "== clean"
	rm $(shell go env GOPATH)/bin/cassandra-operator || true
	rm -rfv $(buildDir)

generate-manifests: setup
	@echo "== generate-manifests"
	$(projectDir)/hack/update-webhooks.sh

validate-manifests: setup
	@echo "== validate-manifests"
	$(projectDir)/hack/verify-webhooks.sh

release: common-docker-release
