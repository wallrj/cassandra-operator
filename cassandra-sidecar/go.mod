module github.com/sky-uk/cassandra-operator/cassandra-sidecar

go 1.12

require (
	github.com/sirupsen/logrus v1.4.2
	github.com/sky-uk/cassandra-operator/cassandra-operator v0.0.0-20190607105530-f2a6996272c3
	github.com/spf13/cobra v0.0.4
	k8s.io/apimachinery v0.0.0-20190606174813-5a6182816fbf
)

replace github.com/sky-uk/cassandra-operator/cassandra-operator => ../cassandra-operator

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190216013122-f05b8decd79c
