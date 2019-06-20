package validation

import (
	"fmt"
	"reflect"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	v1alpha1helpers "github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1/helpers"
)

func ValidateCassandra(c *v1alpha1.Cassandra) field.ErrorList {
	allErrs := apimachineryvalidation.ValidateObjectMeta(
		&c.ObjectMeta,
		true,
		apimachineryvalidation.NameIsDNSSubdomain,
		field.NewPath("metadata"),
	)
	allErrs = append(allErrs, validateCassandraSpec(c, field.NewPath("spec"))...)
	return allErrs
}

func validateCassandraSpec(c *v1alpha1.Cassandra, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validateRacks(c, fldPath.Child("Racks"))...)
	allErrs = append(allErrs, validatePodResources(c, fldPath.Child("Pod"))...)
	return allErrs
}

func validateRacks(clusterDefinition *v1alpha1.Cassandra, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if len(clusterDefinition.Spec.Racks) == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				fldPath,
				fmt.Sprintf(
					"no racks specified for cluster: %s.%s",
					clusterDefinition.Namespace,
					clusterDefinition.Name,
				),
			),
		)
		return allErrs
	}

	for i, rack := range clusterDefinition.Spec.Racks {
		fldPath = fldPath.Child(fmt.Sprintf("%d:%s", i, rack.Name))
		if rack.Replicas < 1 {
			allErrs = append(
				allErrs,
				field.Invalid(
					fldPath.Child("Replicas"),
					rack.Replicas,
					fmt.Sprintf("replicas must be a positive integer for Cassandra cluster: %s.%s",
						clusterDefinition.Namespace,
						clusterDefinition.Name,
					),
				),
			)
		}
		if rack.StorageClass == "" && !v1alpha1helpers.UseEmptyDir(clusterDefinition) {
			allErrs = append(
				allErrs,
				field.Required(
					fldPath.Child("StorageClass"),
					fmt.Sprintf(
						"either set useEmptyDir to true or specify storage class in Cassandra cluster %s.%s",
						clusterDefinition.Namespace,
						clusterDefinition.Name,
					),
				),
			)
		}
		if rack.Zone == "" && !v1alpha1helpers.UseEmptyDir(clusterDefinition) {
			allErrs = append(
				allErrs,
				field.Required(
					fldPath.Child("StorageClass"),
					fmt.Sprintf(
						"either set useEmptyDir to true or specify zone in Cassandra cluster %s.%s",
						clusterDefinition.Namespace,
						clusterDefinition.Name,
					),
				),
			)
		}
	}
	return allErrs
}

func validatePodResources(c *v1alpha1.Cassandra, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if c.Spec.Pod.Memory.IsZero() {
		allErrs = append(
			allErrs,
			field.Invalid(
				fldPath.Child("Memory"),
				c.Spec.Pod.Memory.String(),
				fmt.Sprintf(
					"must be > 0 in Cassandra cluster %s.%s",
					c.Namespace,
					c.Name,
				),
			),
		)
	}
	if !v1alpha1helpers.UseEmptyDir(c) && c.Spec.Pod.StorageSize.IsZero() {
		allErrs = append(
			allErrs,
			field.Invalid(
				fldPath.Child("StorageSize"),
				c.Spec.Pod.StorageSize.String(),
				fmt.Sprintf(
					"must be > 0 when useEmptyDir is false for Cassandra cluster %s.%s",
					c.Namespace,
					c.Name,
				),
			),
		)
	}
	if v1alpha1helpers.UseEmptyDir(c) && !c.Spec.Pod.StorageSize.IsZero() {
		allErrs = append(
			allErrs,
			field.Invalid(
				fldPath.Child("StorageSize"),
				c.Spec.Pod.StorageSize.String(),
				fmt.Sprintf(
					"must be 0 when useEmptyDir is true for Cassandra cluster %s.%s",
					c.Namespace,
					c.Name,
				),
			),
		)
	}
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
					*old.Spec.Pod.Image, *new.Spec.Pod.Image,
				),
			),
		)
	}
	return allErrs
}
