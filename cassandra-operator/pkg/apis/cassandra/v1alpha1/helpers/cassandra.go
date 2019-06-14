package helpers

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
)

func NewControllerRef(c *v1alpha1.Cassandra) metav1.OwnerReference {
	return *metav1.NewControllerRef(c, schema.GroupVersionKind{
		Group:   cassandra.GroupName,
		Version: cassandra.Version,
		Kind:    cassandra.Kind,
	})
}

// UseEmptyDir returns a dereferenced value for Spec.UseEmptyDir
func UseEmptyDir(c *v1alpha1.Cassandra) bool {
	if c.Spec.UseEmptyDir != nil {
		return *c.Spec.UseEmptyDir
	}
	return false
}

// GetImage returns the image for a cluster
func GetCassandraImage(c *v1alpha1.Cassandra) string {
	if c.Spec.Pod.Image != nil {
		return *c.Spec.Pod.Image
	}
	return v1alpha1.DefaultCassandraImage
}

// GetBootstrapperImage returns the bootstrapper image for a cluster
func GetBootstrapperImage(c *v1alpha1.Cassandra) string {
	if c.Spec.Pod.BootstrapperImage != nil {
		return *c.Spec.Pod.BootstrapperImage
	}
	return v1alpha1.DefaultCassandraBootstrapperImage
}

// GetCassandraSidecarImage returns the sidecar image for a cluster
func GetCassandraSidecarImage(c *v1alpha1.Cassandra) string {
	if c.Spec.Pod.SidecarImage != nil {
		return *c.Spec.Pod.SidecarImage
	}
	return v1alpha1.DefaultCassandraSidecarImage
}

// GetSnapshotImage returns the snapshot image for a cluster
func GetSnapshotImage(c *v1alpha1.Cassandra) string {
	if c.Spec.Snapshot != nil {
		if c.Spec.Snapshot.Image != nil {
			return *c.Spec.Snapshot.Image
		}
	}
	return v1alpha1.DefaultCassandraSnapshotImage
}

func GetDatacenter(c *v1alpha1.Cassandra) string {
	if c.Spec.Datacenter == nil {
		return v1alpha1.DefaultDCName
	}
	return *c.Spec.Datacenter
}

// HasRetentionPolicyEnabled returns true when a retention policy exists and is enabled
func HasRetentionPolicyEnabled(snapshot *v1alpha1.Snapshot) bool {
	return snapshot.RetentionPolicy != nil && *snapshot.RetentionPolicy.Enabled
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
