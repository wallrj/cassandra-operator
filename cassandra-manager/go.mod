module github.com/sky-uk/cassandra-operator/cassandra-manager

go 1.12

require (
	github.com/mitchellh/go-homedir v1.0.0 // indirect
	github.com/onsi/ginkgo v1.7.0
	github.com/sky-uk/cassandra-operator/cassandra-operator v0.0.0-00010101000000-000000000000
	github.com/sky-uk/licence-compliance-checker v1.1.0
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20190513172903-22d7a77e9e5f // indirect
	golang.org/x/net v0.0.0-20190522155817-f3200d17e092 // indirect
	golang.org/x/sys v0.0.0-20190523142557-0e01d883c5c5 // indirect
	golang.org/x/text v0.3.2 // indirect
	k8s.io/apimachinery v0.0.0-20190502092502-a44ef629a3c9
)

replace github.com/sky-uk/cassandra-operator/cassandra-operator => ../cassandra-operator

replace k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190216013122-f05b8decd79c
