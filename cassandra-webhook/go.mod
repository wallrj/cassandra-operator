module github.com/sky-uk/cassandra-operator/cassandra-webhook

go 1.12

require (
	github.com/prometheus/client_golang v0.9.4 // indirect
	github.com/sirupsen/logrus v1.3.0
	github.com/sky-uk/cassandra-operator/cassandra-operator v0.0.0-20190613162239-44aa2e756218
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	sigs.k8s.io/controller-runtime v0.0.0-00010101000000-000000000000
)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d

replace github.com/sky-uk/cassandra-operator/cassandra-operator => ../cassandra-operator

replace sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.2.0-beta.2

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.2.0-beta.2
