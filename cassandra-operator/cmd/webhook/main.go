package main

import (
	"github.com/openshift/generic-admission-server/pkg/cmd"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/validation/webhooks"
)

var cassandraValidationHook cmd.ValidatingAdmissionHook = &webhooks.Cassandra{}

func main() {
	cmd.RunAdmissionServer(
		cassandraValidationHook,
	)
}
