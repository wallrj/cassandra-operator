package validation

import (
	"fmt"
	"reflect"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/robfig/cron"
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
				"must not be empty",
			),
		)
		return allErrs
	}

	for i, rack := range c.Spec.Racks {
		fldPath = fldPath.Child(fmt.Sprintf("%d:%s", i, rack.Name))
		if rack.Replicas < 1 {
			allErrs = append(
				allErrs,
				field.Invalid(
					fldPath.Child("Replicas"),
					rack.Replicas,
					"must be > 0",
				),
			)
		}
		if rack.StorageClass == "" && !v1alpha1helpers.UseEmptyDir(c) {
			allErrs = append(
				allErrs,
				field.Required(
					fldPath.Child("StorageClass"),
					"must not be empty if useEmptyDir is true",
				),
			)
		}
		if rack.Zone == "" && !v1alpha1helpers.UseEmptyDir(c) {
			allErrs = append(
				allErrs,
				field.Required(
					fldPath.Child("Zone"),
					"must not be empty if useEmptyDir is true",
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
	if !v1alpha1helpers.UseEmptyDir(c) && c.Spec.Pod.StorageSize.IsZero() {
		allErrs = append(
			allErrs,
			field.Invalid(
				fldPath.Child("StorageSize"),
				c.Spec.Pod.StorageSize.String(),
				"must be > 0 when useEmptyDir is false",
			),
		)
	}
	if v1alpha1helpers.UseEmptyDir(c) && !c.Spec.Pod.StorageSize.IsZero() {
		allErrs = append(
			allErrs,
			field.Invalid(
				fldPath.Child("StorageSize"),
				c.Spec.Pod.StorageSize.String(),
				"must be 0 when useEmptyDir is true",
			),
		)
	}
	allErrs = append(allErrs, validateProbe(c, c.Spec.Pod.LivenessProbe, fldPath.Child("LivenessProbe"))...)
	if *c.Spec.Pod.LivenessProbe.SuccessThreshold != 1 {
		allErrs = append(
			allErrs,
			field.Invalid(
				fldPath.Child("LivenessProbe").Child("SuccessThreshold"),
				*c.Spec.Pod.LivenessProbe.SuccessThreshold,
				"must be 1",
			),
		)
	}
	allErrs = append(allErrs, validateProbe(c, c.Spec.Pod.ReadinessProbe, fldPath.Child("ReadinessProbe"))...)
	return allErrs
}

func validateUnsignedInt(allErrs field.ErrorList, c *v1alpha1.Cassandra, fldPath *field.Path, value int32, min int32) field.ErrorList {
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

func validateProbe(c *v1alpha1.Cassandra, probe *v1alpha1.Probe, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = validateUnsignedInt(allErrs, c, fldPath.Child("FailureThreshold"), *probe.FailureThreshold, 1)
	allErrs = validateUnsignedInt(allErrs, c, fldPath.Child("InitialDelaySeconds"), *probe.InitialDelaySeconds, 0)
	allErrs = validateUnsignedInt(allErrs, c, fldPath.Child("PeriodSeconds"), *probe.PeriodSeconds, 1)
	allErrs = validateUnsignedInt(allErrs, c, fldPath.Child("SuccessThreshold"), *probe.SuccessThreshold, 1)
	allErrs = validateUnsignedInt(allErrs, c, fldPath.Child("TimeoutSeconds"), *probe.TimeoutSeconds, 1)
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
	if c.Spec.Snapshot.TimeoutSeconds != nil {
		allErrs = validateUnsignedInt(allErrs, c, fldPath.Child("TimeoutSeconds"), *c.Spec.Snapshot.TimeoutSeconds, 1)
	}
	allErrs = append(
		allErrs,
		validateSnapshotRetentionPolicy(c, fldPath.Child("RetentionPolicy"))...,
	)
	return allErrs
}

func validateSnapshotRetentionPolicy(c *v1alpha1.Cassandra, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if c.Spec.Snapshot.RetentionPolicy == nil {
		return allErrs
	}
	if c.Spec.Snapshot.RetentionPolicy.RetentionPeriodDays != nil {
		allErrs = validateUnsignedInt(allErrs, c, fldPath.Child("RetentionPeriodDays"), *c.Spec.Snapshot.RetentionPolicy.RetentionPeriodDays, 1)
	}
	if c.Spec.Snapshot.RetentionPolicy.CleanupTimeoutSeconds != nil {
		allErrs = validateUnsignedInt(allErrs, c, fldPath.Child("CleanupTimeoutSeconds"), *c.Spec.Snapshot.RetentionPolicy.CleanupTimeoutSeconds, 1)
	}
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
