.DEFAULT_GOAL := travis
.PHONY: travis clean install check build fmt vet lint validate-licences test e2e-te setup go-dep go-build go-install docker-image push-image release check-style

# Default context is running locally against a dind cluster
KUBE_CONTEXT ?= dind
POD_START_TIMEOUT ?= 120s
USE_MOCK ?= true
FAKE_CASSANDRA_IMAGE ?= $(TEST_REGISTRY)/fake-cassandra:v$(gitRev)
CASSANDRA_BOOTSTRAPPER_IMAGE ?= $(TEST_REGISTRY)/cassandra-bootstrapper:v$(gitRev)
CASSANDRA_SNAPSHOT_IMAGE ?= $(TEST_REGISTRY)/cassandra-snapshot:v$(gitRev)
DOMAIN ?= localhost
IMAGE_TO_DEPLOY ?= $(dockerTestImage)
NAMESPACE ?= test-cassandra-operator
GINKGO_NODES ?= 0
GINKGO_COMPILERS ?= 0

projectDir := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
buildDir := $(projectDir)/build
image := cassandra-operator
junitReportDir := $(buildDir)/junit-reports
pkgs := $(shell go list ./... | grep -v /vendor/ | grep -v /pkg/apis/ | grep -v /pkg/client/ | grep -v /test/e2e)
testPkgs := $(shell go list ./... | grep /test/e2e)

include $(projectDir)/../MakefileDocker

travis: clean install check

build: go-build go-install docker-image

install: build push-image

check-style: fmt vet lint

check: check-style validate-licences test deploy-operator integration-test e2e-test

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
	go build -o $(buildDir)/bin/cassandra-operator -v github.com/sky-uk/cassandra-operator/cassandra-operator/cmd/operator
	KUBE_CONTEXT=ignore go test -run xxxxx $(testPkgs) # build the test code but don't run any tests yet

go-install: go-build
	@echo "== install"
	cp -v $(buildDir)/bin/cassandra-operator $(shell go env GOPATH)/bin/cassandra-operator

fmt:
	go fmt ./...

vet: go-dep
	go vet $(pkgs)

lint:
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
	KUBE_CONTEXT=$(KUBE_CONTEXT) \
	    NAMESPACE=$(NAMESPACE) \
	    POD_START_TIMEOUT=5m \
	    USE_MOCK=false \
	    CASSANDRA_SNAPSHOT_IMAGE=$(CASSANDRA_SNAPSHOT_IMAGE) \
	    CASSANDRA_BOOTSTRAPPER_IMAGE=$(CASSANDRA_BOOTSTRAPPER_IMAGE) \
	    ginkgo -p -nodes=$(GINKGO_NODES) -r -compilers=$(GINKGO_COMPILERS) --v --progress -stream -timeout=30m test/e2e/parallel/creation -- -junit-report-dir $(junitReportDir)
endif

e2e-test: e2e-test-parallel e2e-test-sequential

e2e-test-parallel: ensure-test-report-dir-exists go-dep
	@echo "== e2e-test-parallel"
	KUBE_CONTEXT=$(KUBE_CONTEXT) \
	    NAMESPACE=$(NAMESPACE) \
	    POD_START_TIMEOUT=$(POD_START_TIMEOUT) \
	    FAKE_CASSANDRA_IMAGE=$(FAKE_CASSANDRA_IMAGE) \
	    USE_MOCK=$(USE_MOCK) \
	    CASSANDRA_SNAPSHOT_IMAGE=$(CASSANDRA_SNAPSHOT_IMAGE) \
	    CASSANDRA_BOOTSTRAPPER_IMAGE=$(CASSANDRA_BOOTSTRAPPER_IMAGE) \
	    ginkgo -p -nodes=$(GINKGO_NODES) -r -compilers=$(GINKGO_COMPILERS) --v --progress -stream -timeout=30m test/e2e/parallel -- -junit-report-dir $(junitReportDir)

e2e-test-sequential: ensure-test-report-dir-exists go-dep
	@echo "== e2e-test-sequential"
	KUBE_CONTEXT=$(KUBE_CONTEXT) \
	    NAMESPACE=$(NAMESPACE) \
	    POD_START_TIMEOUT=$(POD_START_TIMEOUT) \
	    FAKE_CASSANDRA_IMAGE=$(FAKE_CASSANDRA_IMAGE) \
	    USE_MOCK=$(USE_MOCK) \
	    CASSANDRA_SNAPSHOT_IMAGE=$(CASSANDRA_SNAPSHOT_IMAGE) \
	    CASSANDRA_BOOTSTRAPPER_IMAGE=$(CASSANDRA_BOOTSTRAPPER_IMAGE) \
	    ginkgo -r -compilers=$(GINKGO_COMPILERS) --v --progress -timeout=30m test/e2e/sequential -- -junit-report-dir $(junitReportDir)

docker-image: go-build common-docker-build

push-image: docker-image common-docker-push

deploy-operator:
	@echo "== deploy-operator"
	IMAGE=$(IMAGE_TO_DEPLOY) CONTEXT=$(KUBE_CONTEXT) NAMESPACE=$(NAMESPACE) INGRESS_HOST=$(NAMESPACE).cassandra-operator.$(DOMAIN) $(projectDir)/deploy.sh

validate-licences:
	@echo "== validate licences"
	set -e ;\
	restricted=$$(paste -s -d ',' restricted-licences.txt) ;\
	projects=$$(dep status -f='vendor/{{ .ProjectRoot }} ') ;\
	licence-compliance-checker -L error -E -r $$restricted -o vendor/sigs.k8s.io/yaml=BSD-3-Clause -o vendor/github.com/golang/protobuf=BSD-3-Clause -o vendor/github.com/gogo/protobuf=BSD-3-Clause -o vendor/github.com/hashicorp/golang-lru=MPL-2.0 $$projects ;\

clean: common-docker-clean
	@echo "== clean"
	rm $(shell go env GOPATH)/bin/cassandra-operator || true
	rm -rfv $(buildDir)

generate-types:
	@echo "== generate-types"
	$(projectDir)/hack/update-codegen.sh

release: common-docker-release
