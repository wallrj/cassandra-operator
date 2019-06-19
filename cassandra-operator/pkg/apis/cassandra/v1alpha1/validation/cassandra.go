package validation

import (
	"fmt"
	"reflect"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
)

func ValidateCassandra(c *v1alpha1.Cassandra) field.ErrorList {
	allErrs := apimachineryvalidation.ValidateObjectMeta(&c.ObjectMeta, true, apimachineryvalidation.NameIsDNSSubdomain, field.NewPath("metadata"))
	// allErrs = append(allErrs, ValidateCassandraClusterSpec(&c.Spec, field.NewPath("spec"))...)
	return allErrs
}

func ValidateCassandraUpdate(old, new *v1alpha1.Cassandra) field.ErrorList {
	allErrs := ValidateCassandra(new)

	fldPath := field.NewPath("spec")
	if !reflect.DeepEqual(new.Spec.Pod.Image, old.Spec.Pod.Image) {
		allErrs = append(
			allErrs,
			field.Forbidden(
				fldPath.Child("pod").Child("image"),
				fmt.Sprintf(
					"cannot change the image. "+
						"old version: %s, new version: %s",
					old.Spec.Pod.Image, new.Spec.Pod.Image,
				),
			),
		)
	}
	return allErrs
}
