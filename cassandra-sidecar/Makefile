.DEFAULT_GOAL := travis
.PHONY: travis clean install check build fmt vet lint validate-licences test e2e-te setup go-build go-install docker-image push-image release check-style

# Default context is running locally against a dind cluster
KUBE_CONTEXT ?= dind
USE_MOCK ?= false
FAKE_CASSANDRA_IMAGE ?= $(TEST_REGISTRY)/fake-cassandra:v$(gitRev)
IMAGE_UNDER_TEST ?= $(dockerTestImage)
NAMESPACE ?= test-cassandra-operator
GINKGO_NODES ?= 0
GINKGO_COMPILERS ?= 0

projectDir := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
buildDir := $(projectDir)/build
image := cassandra-sidecar
junitReportDir = $(buildDir)/junit-reports
pkgs := $(shell go list ./... | grep -v /vendor/ | grep -v /test/e2e)
testPkgs := $(shell go list ./... | grep /test/e2e)

include $(projectDir)/../MakefileDocker

travis: clean install check

build: go-build docker-image

install: build go-install push-image

check-style: fmt vet lint

check: check-style validate-licences test integration-test e2e-test

ensure-go-modules:
ifeq ('$(shell go env | grep -o "GOMOD.*")', 'GOMOD=""')
	$(error "go modules must be enabled: export GO111MODULE=on, or run from outside $$GOPATH")
else
	@echo "go modules enabled"
endif

setup:
	@echo "== setup"
	go install \
		github.com/sky-uk/licence-compliance-checker \
		github.com/onsi/ginkgo/ginkgo \
		golang.org/x/tools/cmd/goimports \
		golang.org/x/lint/golint

go-build: ensure-build-dir-exists ensure-go-modules
	@echo "== build"
	go build -o $(buildDir)/bin/cassandra-sidecar -v github.com/sky-uk/cassandra-operator/cassandra-sidecar/cmd
	# FIXME: add some tests
	# KUBE_CONTEXT=ignore go test -run xxxxx $(testPkgs)  # build the test code but don't run any tests yet

go-install:
	@echo "== go-install"
	cp -v $(buildDir)/bin/cassandra-sidecar $(shell go env GOPATH)/bin/cassandra-sidecar

fmt:
	go fmt ./...

validate-licences:
	@echo "== validate licences"
	set -e ;\
	restricted=$$(paste -s -d ',' restricted-licences.txt) ;\
	licence-compliance-checker -L error -E -r $$restricted \
		-m github.com/sky-uk/cassandra-operator/cassandra-sidecar=BSD-3-Clause \
		-m sigs.k8s.io/yaml=BSD-3-Clause \
		--check-go-modules ;\

vet:
	@echo "== vet"
	go vet $(pkgs)

lint:
	@echo "== lint"
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

clean: common-docker-clean
	@echo "== clean"
	rm $(shell go env GOPATH)/bin/cassandra-sidecar || true
	rm -rfv $(buildDir)

release: common-docker-release
