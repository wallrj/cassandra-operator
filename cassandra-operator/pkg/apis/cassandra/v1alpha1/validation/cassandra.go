package validation

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/robfig/cron"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
)

// ValidateCassandra checks that all required fields are supplied and that they have valid values
// NB ObjectMeta is not validated here;
// apiVersion, kind and metadata, are all validated by the API server implicitly.
// See https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#specifying-a-structural-schema
func ValidateCassandra(c *v1alpha1.Cassandra) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validateCassandraSpec(c, field.NewPath("spec"))...)
	return allErrs
}

func validateCassandraSpec(c *v1alpha1.Cassandra, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validateRacks(c, fldPath.Child("Racks"))...)
	allErrs = append(allErrs, validatePodResources(c, fldPath.Child("Pod"))...)
	allErrs = append(allErrs, validateSnapshot(c, fldPath.Child("Snapshot"))...)
	return allErrs
}

func validateRacks(c *v1alpha1.Cassandra, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if len(c.Spec.Racks) == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				fldPath,
				"",
			),
		)
		return allErrs
	}

	useEmptyDir := *c.Spec.UseEmptyDir
	for _, rack := range c.Spec.Racks {
		fldPath = fldPath.Child(rack.Name)
		allErrs = validateUnsignedInt(allErrs, fldPath.Child("Replicas"), rack.Replicas, 1)
		if rack.StorageClass == "" && !useEmptyDir {
			allErrs = append(
				allErrs,
				field.Required(
					fldPath.Child("StorageClass"),
					"because spec.useEmptyDir is false",
				),
			)
		}
		if rack.Zone == "" && !useEmptyDir {
			allErrs = append(
				allErrs,
				field.Required(
					fldPath.Child("Zone"),
					"because spec.useEmptyDir is false",
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
				"must be > 0",
			),
		)
	}
	if !*c.Spec.UseEmptyDir && c.Spec.Pod.StorageSize.IsZero() {
		allErrs = append(
			allErrs,
			field.Invalid(
				fldPath.Child("StorageSize"),
				c.Spec.Pod.StorageSize.String(),
				"must be > 0 when spec.useEmptyDir is false",
			),
		)
	}
	if *c.Spec.UseEmptyDir && !c.Spec.Pod.StorageSize.IsZero() {
		allErrs = append(
			allErrs,
			field.Invalid(
				fldPath.Child("StorageSize"),
				c.Spec.Pod.StorageSize.String(),
				"must be 0 when spec.useEmptyDir is true",
			),
		)
	}

	allErrs = append(
		allErrs,
		validateLivenessProbe(c.Spec.Pod.LivenessProbe, fldPath.Child("LivenessProbe"))...,
	)
	allErrs = append(
		allErrs,
		validateReadinessProbe(c.Spec.Pod.ReadinessProbe, fldPath.Child("ReadinessProbe"))...,
	)
	return allErrs
}

func validateUnsignedInt(allErrs field.ErrorList, fldPath *field.Path, value int32, min int32) field.ErrorList {
	if value < min {
		allErrs = append(
			allErrs,
			field.Invalid(
				fldPath,
				value,
				fmt.Sprintf("must be >= %d", min),
			),
		)
	}
	return allErrs
}

// validateLivenessProbe wraps `validateProbe` and filters out the results for `SuccessThreshold`,
// instead performing a LivenessProbe specific check, to ensure that the value is always 1.
// This is explained in the Kubernetes API docs as follows:
//   Minimum consecutive successes for the probe to be considered successful after having failed.
//   Defaults to 1. Must be 1 for liveness. Minimum value is 1.
// See https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#probe-v1-core
func validateLivenessProbe(probe *v1alpha1.Probe, fldPath *field.Path) field.ErrorList {
	allErrs := validateProbe(probe, fldPath)
	successThresholdFieldPath := fldPath.Child("SuccessThreshold")
	allErrs = allErrs.Filter(func(e error) bool {
		fieldErr, ok := e.(*field.Error)
		if ok && fieldErr.Field == successThresholdFieldPath.String() {
			return true
		}
		return false
	})
	if *probe.SuccessThreshold != 1 {
		allErrs = append(
			allErrs,
			field.Invalid(
				successThresholdFieldPath,
				*probe.SuccessThreshold,
				"must be 1",
			),
		)
	}
	return allErrs
}

func validateReadinessProbe(probe *v1alpha1.Probe, fldPath *field.Path) field.ErrorList {
	return validateProbe(probe, fldPath)
}

func validateProbe(probe *v1alpha1.Probe, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = validateUnsignedInt(allErrs, fldPath.Child("FailureThreshold"), *probe.FailureThreshold, 1)
	allErrs = validateUnsignedInt(allErrs, fldPath.Child("InitialDelaySeconds"), *probe.InitialDelaySeconds, 0)
	allErrs = validateUnsignedInt(allErrs, fldPath.Child("PeriodSeconds"), *probe.PeriodSeconds, 1)
	allErrs = validateUnsignedInt(allErrs, fldPath.Child("SuccessThreshold"), *probe.SuccessThreshold, 1)
	allErrs = validateUnsignedInt(allErrs, fldPath.Child("TimeoutSeconds"), *probe.TimeoutSeconds, 1)
	return allErrs
}

func validateSnapshot(c *v1alpha1.Cassandra, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if c.Spec.Snapshot == nil {
		return allErrs
	}
	if _, err := cron.Parse(c.Spec.Snapshot.Schedule); err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(
				fldPath.Child("Schedule"),
				c.Spec.Snapshot.Schedule,
				fmt.Sprintf(
					"is not a valid cron expression (%s)",
					err,
				),
			),
		)
	}
	allErrs = validateUnsignedInt(allErrs, fldPath.Child("TimeoutSeconds"), *c.Spec.Snapshot.TimeoutSeconds, 1)
	if c.Spec.Snapshot.RetentionPolicy != nil {
		allErrs = append(
			allErrs,
			validateSnapshotRetentionPolicy(c, fldPath.Child("RetentionPolicy"))...,
		)
	}
	return allErrs
}

func validateSnapshotRetentionPolicy(c *v1alpha1.Cassandra, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = validateUnsignedInt(allErrs, fldPath.Child("RetentionPeriodDays"), *c.Spec.Snapshot.RetentionPolicy.RetentionPeriodDays, 0)
	allErrs = validateUnsignedInt(allErrs, fldPath.Child("CleanupTimeoutSeconds"), *c.Spec.Snapshot.RetentionPolicy.CleanupTimeoutSeconds, 0)
	if _, err := cron.Parse(c.Spec.Snapshot.RetentionPolicy.CleanupSchedule); err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(
				fldPath.Child("CleanupSchedule"),
				c.Spec.Snapshot.RetentionPolicy.CleanupSchedule,
				fmt.Sprintf(
					"is not a valid cron expression (%s)",
					err,
				),
			),
		)
	}
	return allErrs
}
