module github.com/sky-uk/cassandra-operator/cassandra-webhook

require (
	github.com/evanphx/json-patch v4.2.0+incompatible // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/spf13/pflag v1.0.3 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/net v0.0.0-20190501004415-9ce7a6920f09 // indirect
	golang.org/x/oauth2 v0.0.0-20181203162652-d668ce993890 // indirect
	golang.org/x/sys v0.0.0-20190509141414-a5b02f93d862 // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	k8s.io/api v0.0.0-20190503110853-61630f889b3c // indirect
	sigs.k8s.io/controller-runtime v0.1.12
)

replace github.com/openshift/generic-admission-server => github.com/openshift/generic-admission-server v1.14.0

replace github.com/sky-uk/cassandra-operator/cassandra-operator => ../cassandra-operator

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190413052642-108c485f896e

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.2.0-beta.2

replace sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.2.0-beta.2
