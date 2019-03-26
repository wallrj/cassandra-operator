.DEFAULT_GOAL := travis
.PHONY: travis clean install check build fmt vet lint validate-licences test e2e-te setup go-build go-install docker-image push-image

# Default context is running locally against a dind cluster
KUBE_CONTEXT ?= dind
USE_MOCK ?= false
FAKE_CASSANDRA_IMAGE ?= $(TEST_REGISTRY)/fake-cassandra:v$(gitRev)
TEST_REGISTRY ?= localhost:5000

projectDir := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
k8ResourcesDir := $(projectDir)/../cassandra-operator/kubernetes-resources
buildDir := $(projectDir)/build
gitRev := $(shell git rev-parse --short HEAD)
junitReportDir = $(buildDir)/junit-reports
pkgs := $(shell go list ./... | grep -v /vendor/ | grep -v /test/e2e)
testPkgs := $(shell go list ./... | grep /test/e2e)
testRunOptions := IMAGE_UNDER_TEST=$(TEST_REGISTRY)/cassandra-snapshot:v$(gitRev) FAKE_CASSANDRA_IMAGE=$(FAKE_CASSANDRA_IMAGE) KUBE_CONTEXT=$(KUBE_CONTEXT) USE_MOCK=$(USE_MOCK)

travis: clean install check

build: go-build docker-image

install: build go-install push-image

check: fmt vet lint validate-licences test e2e-test

setup:
	@echo "== setup"
	go get -v golang.org/x/lint/golint golang.org/x/tools/cmd/goimports github.com/golang/dep/cmd/dep github.com/sky-uk/licence-compliance-checker
	go get -v github.com/onsi/ginkgo/ginkgo && cd $$GOPATH/src/github.com/onsi/ginkgo && git checkout 'v1.6.0' && go install github.com/onsi/ginkgo/ginkgo
	curl https://raw.githubusercontent.com/golang/dep/v0.5.1/install.sh | sh

go-build: ensure-build-dir-exists
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

test: ensure-test-report-dir-exists
	@echo "== test"
	ginkgo -r --v --progress pkg cmd -- -junit-report-dir $(junitReportDir)

e2e-test: ensure-test-report-dir-exists push-image kubernetes-resources
	@echo "== e2e-test"
	$(testRunOptions) ginkgo -r --v --progress test/e2e -- -junit-report-dir $(junitReportDir)

kubernetes-resources:
	@echo "== kubernetes-resources"
	sed -e "s@\$$TARGET_NAMESPACE@test-cassandra-operator@g" $(k8ResourcesDir)/cassandra-snapshot.yml > ${buildDir}/cassandra-snapshot.yml
	kubectl --context $(KUBE_CONTEXT) -n test-cassandra-operator apply -f ${buildDir}/cassandra-snapshot.yml

docker-image: go-build
	@echo "== docker-image"
	@echo "Building image v$(gitRev)"
	docker build . -t $(TEST_REGISTRY)/cassandra-snapshot:v$(gitRev)

push-image: docker-image
	@echo "== push-image"
	docker push $(TEST_REGISTRY)/cassandra-snapshot:v$(gitRev)
#	@docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)
#	docker tag $(TEST_REGISTRY)/cassandra-snapshot:v$(gitRev) skyuk/cassandra-operator:cassandra-snapshot-v$(gitRev)
#	docker push skyuk/cassandra-operator:cassandra-snapshot-v$(gitRev)

clean:
	@echo "== clean"
	docker rm $(TEST_REGISTRY)/cassandra-snapshot:v$(gitRev) || true
	rm $(shell go env GOPATH)/bin/cassandra-snapshot || true
	rm -rfv $(buildDir)
