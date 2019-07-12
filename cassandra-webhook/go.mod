module github.com/sky-uk/cassandra-operator/cassandra-webhook

go 1.12

require (
	github.com/GeertJohan/go.rice v1.0.0 // indirect
	github.com/cloudflare/cfssl v0.0.0-20190627231140-2001f384ec4f
	github.com/go-sql-driver/mysql v1.4.1 // indirect
	github.com/google/certificate-transparency-go v1.0.21 // indirect
	github.com/jmhodges/clock v0.0.0-20160418191101-880ee4c33548 // indirect
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/kisielk/sqlstruct v0.0.0-20150923205031-648daed35d49 // indirect
	github.com/lib/pq v1.1.1 // indirect
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/prometheus/client_golang v0.9.4 // indirect
	github.com/sirupsen/logrus v1.3.0
	github.com/sky-uk/cassandra-operator/cassandra-operator v0.0.0-20190613162239-44aa2e756218
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apimachinery v0.0.0-20190502092502-a44ef629a3c9
	sigs.k8s.io/controller-runtime v0.0.0-00010101000000-000000000000
	sigs.k8s.io/yaml v1.1.0
)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d

replace github.com/sky-uk/cassandra-operator/cassandra-operator => ../cassandra-operator

replace sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.2.0-beta.3

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.2.0-beta.3
