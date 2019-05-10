package main

import (
	"github.com/openshift/generic-admission-server/pkg/cmd"
	"github.com/sky-uk/cassandra-operator/cassandra-webhook/pkg/apis/cassandra/validation/webhooks"
)

var hook cmd.ValidatingAdmissionHook = &webhooks.Cassandra{}

func main() {
	cmd.RunAdmissionServer(hook)
}
