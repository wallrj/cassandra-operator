package validation

import (
	"fmt"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/robfig/cron"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
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
					"because spec.useEmptyDir is true",
				),
			)
		}
		if rack.Zone == "" && !useEmptyDir {
			allErrs = append(
				allErrs,
				field.Required(
					fldPath.Child("Zone"),
					"because spec.useEmptyDir is true",
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
		validateProbe(c.Spec.Pod.LivenessProbe, fldPath.Child("LivenessProbe"), true)...,
	)
	allErrs = append(
		allErrs,
		validateProbe(c.Spec.Pod.ReadinessProbe, fldPath.Child("ReadinessProbe"), false)...,
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

func validateProbe(probe *v1alpha1.Probe, fldPath *field.Path, livenessProbe bool) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = validateUnsignedInt(allErrs, fldPath.Child("FailureThreshold"), *probe.FailureThreshold, 1)
	allErrs = validateUnsignedInt(allErrs, fldPath.Child("InitialDelaySeconds"), *probe.InitialDelaySeconds, 0)
	allErrs = validateUnsignedInt(allErrs, fldPath.Child("PeriodSeconds"), *probe.PeriodSeconds, 1)
	if livenessProbe {
		if *probe.SuccessThreshold != 1 {
			allErrs = append(
				allErrs,
				field.Invalid(
					fldPath.Child("SuccessThreshold"),
					*probe.SuccessThreshold,
					"must be 1",
				),
			)
		}
	} else {
		allErrs = validateUnsignedInt(allErrs, fldPath.Child("SuccessThreshold"), *probe.SuccessThreshold, 1)
	}

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
