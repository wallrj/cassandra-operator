package webhooks

import (
	"github.com/openshift/generic-admission-server/pkg/cmd"
)

type Cassandra struct {
}

var _ cmd.ValidatingAdmissionHook = &Cassandra{}
