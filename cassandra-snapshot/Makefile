.DEFAULT_GOAL := travis
.PHONY: travis clean install check build fmt vet lint validate-licences test e2e-te setup go-dep go-build go-install docker-image push-image prepare-deployment release

# Default context is running locally against a dind cluster
KUBE_CONTEXT ?= dind
USE_MOCK ?= true
FAKE_CASSANDRA_IMAGE ?= $(TEST_REGISTRY)/fake-cassandra:v$(gitRev)
IMAGE_UNDER_TEST ?= $(dockerTestImage)
NAMESPACE ?= test-cassandra-operator
GINKGO_NODES ?= 0
GINKGO_COMPILERS ?= 0

projectDir := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
k8ResourcesDir := $(projectDir)/../cassandra-operator/kubernetes-resources
buildDir := $(projectDir)/build
image := cassandra-snapshot
junitReportDir = $(buildDir)/junit-reports
pkgs := $(shell go list ./... | grep -v /vendor/ | grep -v /test/e2e)
testPkgs := $(shell go list ./... | grep /test/e2e)

include $(projectDir)/../MakefileDocker

travis: clean install check

build: go-build docker-image

install: build go-install push-image

check: fmt vet lint validate-licences test prepare-deployment integration-test e2e-test

setup:
	@echo "== setup"
	go get -v golang.org/x/lint/golint golang.org/x/tools/cmd/goimports github.com/golang/dep/cmd/dep github.com/sky-uk/licence-compliance-checker
	go get -v github.com/onsi/ginkgo/ginkgo && cd $$GOPATH/src/github.com/onsi/ginkgo && git checkout 'v1.6.0' && go install github.com/onsi/ginkgo/ginkgo
	curl https://raw.githubusercontent.com/golang/dep/v0.5.1/install.sh | sh

go-dep:
	@echo "== go-dep"
	dep ensure -v -vendor-only

go-build: ensure-build-dir-exists go-dep
	@echo "== build"
	dep ensure -v -vendor-only
	go build -o $(buildDir)/bin/cassandra-snapshot -v github.com/sky-uk/cassandra-operator/cassandra-snapshot/cmd
	KUBE_CONTEXT=ignore go test -run xxxxx $(testPkgs)  # build the test code but don't run any tests yet

go-install:
	@echo "== go-install"
	cp -v $(buildDir)/bin/cassandra-snapshot $(shell go env GOPATH)/bin/cassandra-snapshot

fmt:
	go fmt ./...

validate-licences:
	@echo "== validate licences"
	set -e ;\
	restricted=$$(paste -s -d ',' restricted-licences.txt) ;\
	projects=$$(dep status -f='vendor/{{ .ProjectRoot }} ') ;\
	licence-compliance-checker -L error -E -r $$restricted -o vendor/github.com/ghodss/yaml=BSD-3-Clause $$projects ;\

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

test: ensure-test-report-dir-exists go-dep
	@echo "== test"
	ginkgo -p -nodes=$(GINKGO_NODES) -r -compilers=$(GINKGO_COMPILERS) --v --progress pkg cmd -- -junit-report-dir $(junitReportDir)

integration-test: ensure-test-report-dir-exists go-dep
	@echo "== integration-test"
ifeq ($(strip $(USE_MOCK)),true)
	@echo "== e2e-test integration test against real cassandra"
	IMAGE_UNDER_TEST=$(dockerTestImage) \
        KUBE_CONTEXT=$(KUBE_CONTEXT) \
        NAMESPACE=$(NAMESPACE) \
        USE_MOCK=false \
        ginkgo -r -compilers=$(GINKGO_COMPILERS) --v --progress test/e2e/create -- -junit-report-dir $(junitReportDir)
endif

e2e-test: ensure-test-report-dir-exists go-dep
	@echo "== e2e-test"
	IMAGE_UNDER_TEST=$(IMAGE_UNDER_TEST) \
        FAKE_CASSANDRA_IMAGE=$(FAKE_CASSANDRA_IMAGE) \
        KUBE_CONTEXT=$(KUBE_CONTEXT) \
        NAMESPACE=$(NAMESPACE) \
        USE_MOCK=$(USE_MOCK) \
        ginkgo -r -compilers=$(GINKGO_COMPILERS) --v --progress test/e2e -- -junit-report-dir $(junitReportDir)

docker-image: go-build common-docker-build

push-image: docker-image common-docker-push

prepare-deployment:
	@echo "== prepare-deployment"
	sed -e "s@\$$TARGET_NAMESPACE@$(NAMESPACE)@g" $(k8ResourcesDir)/cassandra-snapshot.yml > ${buildDir}/cassandra-snapshot.yml
	kubectl --context $(KUBE_CONTEXT) -n $(NAMESPACE) apply -f ${buildDir}/cassandra-snapshot.yml

clean: common-docker-clean
	@echo "== clean"
	rm $(shell go env GOPATH)/bin/cassandra-snapshot || true
	rm -rfv $(buildDir)

release: common-docker-release
