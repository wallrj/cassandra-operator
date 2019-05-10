package validation

import (
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateCassandra(c *v1alpha1.Cassandra) field.ErrorList {
	el := field.ErrorList{}
	return el
}
