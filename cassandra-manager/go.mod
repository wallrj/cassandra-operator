module github.com/sky-uk/cassandra-operator/cassandra-manager

go 1.12

require (
	github.com/onsi/ginkgo v1.7.0
	github.com/sky-uk/cassandra-operator/cassandra-operator v0.0.0-00010101000000-000000000000
	github.com/sky-uk/licence-compliance-checker v1.1.0
)

replace github.com/sky-uk/cassandra-operator/cassandra-operator => ../cassandra-operator

replace k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190216013122-f05b8decd79c
