package helpers

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/util/ptr"
)

func NewControllerRef(c *v1alpha1.Cassandra) metav1.OwnerReference {
	return *metav1.NewControllerRef(c, schema.GroupVersionKind{
		Group:   cassandra.GroupName,
		Version: cassandra.Version,
		Kind:    cassandra.Kind,
	})
}

// HasRetentionPolicyEnabled returns true when a retention policy exists and is enabled
func HasRetentionPolicyEnabled(snapshot *v1alpha1.Snapshot) bool {
	return snapshot.RetentionPolicy != nil &&
		snapshot.RetentionPolicy.Enabled != nil &&
		*snapshot.RetentionPolicy.Enabled
}

// SnapshotPropertiesUpdated returns false when snapshot1 and snapshot2 have the same properties disregarding retention policy
func SnapshotPropertiesUpdated(snapshot1 *v1alpha1.Snapshot, snapshot2 *v1alpha1.Snapshot) bool {
	return snapshot1.Schedule != snapshot2.Schedule ||
		!reflect.DeepEqual(snapshot1.TimeoutSeconds, snapshot2.TimeoutSeconds) ||
		!reflect.DeepEqual(snapshot1.Keyspaces, snapshot2.Keyspaces)
}

// SnapshotCleanupPropertiesUpdated returns false snapshot1 and snapshot2 have the same retention policy regardless of whether it is enabled or not
func SnapshotCleanupPropertiesUpdated(snapshot1 *v1alpha1.Snapshot, snapshot2 *v1alpha1.Snapshot) bool {
	return snapshot1.RetentionPolicy != nil && snapshot2.RetentionPolicy != nil &&
		(snapshot1.RetentionPolicy.CleanupSchedule != snapshot2.RetentionPolicy.CleanupSchedule ||
			!reflect.DeepEqual(snapshot1.RetentionPolicy.CleanupTimeoutSeconds, snapshot2.RetentionPolicy.CleanupTimeoutSeconds) ||
			!reflect.DeepEqual(snapshot1.RetentionPolicy.RetentionPeriodDays, snapshot2.RetentionPolicy.RetentionPeriodDays))
}

func SetDefaultsForCassandra(clusterDefinition *v1alpha1.Cassandra) {
	setDefaultsForDatacenter(clusterDefinition)
	setDefaultsForUseEmptyDir(clusterDefinition)
	setDefaultsForSnapshot(clusterDefinition.Spec.Snapshot)
	setDefaultsForImages(clusterDefinition)
	setDefaultsForProbes(clusterDefinition)
}

func setDefaultsForDatacenter(clusterDefinition *v1alpha1.Cassandra) {
	if clusterDefinition.Spec.Datacenter == nil {
		clusterDefinition.Spec.Datacenter = ptr.String(v1alpha1.DefaultDatacenterName)
	}
}

func setDefaultsForUseEmptyDir(clusterDefinition *v1alpha1.Cassandra) {
	if clusterDefinition.Spec.UseEmptyDir == nil {
		clusterDefinition.Spec.UseEmptyDir = ptr.Bool(false)
	}
}

func setDefaultsForSnapshot(snapshot *v1alpha1.Snapshot) {
	if snapshot == nil {
		return
	}
	if snapshot.TimeoutSeconds == nil {
		snapshot.TimeoutSeconds = ptr.Int32(v1alpha1.DefaultSnapshotTimeoutSeconds)
	}

	if snapshot.RetentionPolicy != nil {
		setDefaultsForRetentionPolicy(snapshot.RetentionPolicy)
	}
}

func setDefaultsForRetentionPolicy(rp *v1alpha1.RetentionPolicy) {
	if rp.Enabled == nil {
		rp.Enabled = ptr.Bool(true)
	}
	if rp.RetentionPeriodDays == nil {
		rp.RetentionPeriodDays = ptr.Int32(v1alpha1.DefaultRetentionPolicyRetentionPeriodDays)
	}
	if rp.CleanupTimeoutSeconds == nil {
		rp.CleanupTimeoutSeconds = ptr.Int32(v1alpha1.DefaultRetentionPolicyCleanupTimeoutSeconds)
	}
}

func setDefaultsForImages(clusterDefinition *v1alpha1.Cassandra) {
	if clusterDefinition.Spec.Pod.BootstrapperImage == nil {
		clusterDefinition.Spec.Pod.BootstrapperImage = ptr.String(v1alpha1.DefaultCassandraBootstrapperImage)
	}
	if clusterDefinition.Spec.Pod.Image == nil {
		clusterDefinition.Spec.Pod.Image = ptr.String(v1alpha1.DefaultCassandraImage)
	}
	if clusterDefinition.Spec.Pod.SidecarImage == nil {
		clusterDefinition.Spec.Pod.SidecarImage = ptr.String(v1alpha1.DefaultCassandraSidecarImage)
	}
	if clusterDefinition.Spec.Snapshot != nil && clusterDefinition.Spec.Snapshot.Image == nil {
		clusterDefinition.Spec.Snapshot.Image = ptr.String(v1alpha1.DefaultCassandraSnapshotImage)
	}
}

func setDefaultsForProbes(clusterDefinition *v1alpha1.Cassandra) {
	if clusterDefinition.Spec.Pod.LivenessProbe == nil {
		defaultProbe := defaultLivenessProbe()
		clusterDefinition.Spec.Pod.LivenessProbe = &defaultProbe
	} else {
		livenessProbe := clusterDefinition.Spec.Pod.LivenessProbe
		mergeProbeDefaults(livenessProbe, defaultLivenessProbe())
	}

	if clusterDefinition.Spec.Pod.ReadinessProbe == nil {
		defaultProbe := defaultReadinessProbe()
		clusterDefinition.Spec.Pod.ReadinessProbe = &defaultProbe
	} else {
		readinessProbe := clusterDefinition.Spec.Pod.ReadinessProbe
		mergeProbeDefaults(readinessProbe, defaultReadinessProbe())
	}
}

func mergeProbeDefaults(configuredProbe *v1alpha1.Probe, defaultProbe v1alpha1.Probe) {
	if configuredProbe.TimeoutSeconds == nil {
		configuredProbe.TimeoutSeconds = defaultProbe.TimeoutSeconds
	}

	if configuredProbe.SuccessThreshold == nil {
		configuredProbe.SuccessThreshold = defaultProbe.SuccessThreshold
	}

	if configuredProbe.FailureThreshold == nil {
		configuredProbe.FailureThreshold = defaultProbe.FailureThreshold
	}

	if configuredProbe.InitialDelaySeconds == nil {
		configuredProbe.InitialDelaySeconds = defaultProbe.InitialDelaySeconds
	}

	if configuredProbe.PeriodSeconds == nil {
		configuredProbe.PeriodSeconds = defaultProbe.PeriodSeconds
	}
}

func defaultReadinessProbe() v1alpha1.Probe {
	return v1alpha1.Probe{
		FailureThreshold:    ptr.Int32(3),
		InitialDelaySeconds: ptr.Int32(30),
		PeriodSeconds:       ptr.Int32(15),
		SuccessThreshold:    ptr.Int32(1),
		TimeoutSeconds:      ptr.Int32(5),
	}
}

func defaultLivenessProbe() v1alpha1.Probe {
	return v1alpha1.Probe{
		FailureThreshold:    ptr.Int32(3),
		InitialDelaySeconds: ptr.Int32(30),
		PeriodSeconds:       ptr.Int32(30),
		SuccessThreshold:    ptr.Int32(1),
		TimeoutSeconds:      ptr.Int32(5),
	}
}
