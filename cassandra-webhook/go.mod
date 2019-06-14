module github.com/sky-uk/cassandra-operator/cassandra-webhook

go 1.12

require (
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/prometheus/client_golang v0.9.4 // indirect
	github.com/sky-uk/cassandra-operator/cassandra-operator v0.0.0-20190613162239-44aa2e756218
	go.etcd.io/etcd v3.3.13+incompatible
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	k8s.io/api v0.0.0-20190612125737-db0771252981 // indirect
	k8s.io/apimachinery v0.0.0-20190612125636-6a5db36e93ad
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a // indirect
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
