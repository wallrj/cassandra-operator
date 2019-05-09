package v1alpha1

import (
	"fmt"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// required for dep management
	"k8s.io/apimachinery/pkg/api/resource"
	_ "k8s.io/code-generator/cmd/client-gen/types"
)

const (
	NodeServiceAccountName     = "cassandra-node"
	SnapshotServiceAccountName = "cassandra-snapshot"
	// DefaultDCName is the default data center name which each Cassandra pod belongs to
	DefaultDCName = "dc1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cassandra defines a Cassandra cluster
type Cassandra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraSpec   `json:"spec"`
	Status CassandraStatus `json:"status"`
}

// CassandraSpec is the specification for the Cassandra resource
type CassandraSpec struct {
	// +optional
	Datacenter *string `json:"datacenter,omitempty"`
	// +optional
	Racks []Rack `json:"racks,omitempty"`
	// +optional
	UseEmptyDir bool `json:"useEmptyDir"`
	Pod         Pod  `json:"pod"`
	// +optional
	Snapshot *Snapshot `json:"snapshot,omitempty"`
}

type Probe struct {
	// +optional
	FailureThreshold int32 `json:"failureThreshold"`
	// +optional
	InitialDelaySeconds int32 `json:"initialDelaySeconds"`
	// +optional
	PeriodSeconds int32 `json:"periodSeconds"`
	// +optional
	SuccessThreshold int32 `json:"successThreshold"`
	// +optional
	TimeoutSeconds int32 `json:"timeoutSeconds"`
}

type Pod struct {
	// +optional
	BootstrapperImage string `json:"bootstrapperImage"`
	// +optional
	Image       string            `json:"image"`
	StorageSize resource.Quantity `json:"storageSize"`
	Memory      resource.Quantity `json:"memory"`
	CPU         resource.Quantity `json:"cpu"`
	// +optional
	LivenessProbe *Probe `json:"livenessProbe"`
	// +optional
	ReadinessProbe *Probe `json:"readinessProbe"`
}

// CassandraStatus is the status for the Cassandra resource
type CassandraStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraList is a list of Cassandra resources
type CassandraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Cassandra `json:"items"`
}

// Rack defines the properties of a rack in the cluster
type Rack struct {
	Name         string `json:"name"`
	Zone         string `json:"zone"`
	StorageClass string `json:"storageClass"`
	Replicas     int32  `json:"replicas"`
}

// Snapshot defines the snapshot creation and deletion configuration
type Snapshot struct {
	// +optional
	Image string `json:"image"`
	// Schedule follows the cron format, see https://en.wikipedia.org/wiki/Cron
	Schedule string `json:"schedule"`
	// +optional
	Keyspaces []string `json:"keyspaces,omitempty"`
	// +optional
	TimeoutSeconds *int32 `json:"timeoutSeconds,omitempty"`
	// +optional
	RetentionPolicy *RetentionPolicy `json:"retentionPolicy,omitempty"`
}

// HasRetentionPolicyEnabled returns true when a retention policy exists and is enabled
func (s *Snapshot) HasRetentionPolicyEnabled() bool {
	return s.RetentionPolicy != nil && s.RetentionPolicy.Enabled
}

// RetentionPolicy defines how long the snapshots should be kept for and how often the cleanup task should run
type RetentionPolicy struct {
	Enabled bool `json:"enabled"`
	// +optional
	RetentionPeriodDays *int32 `json:"retentionPeriodDays,omitempty"`
	// CleanupSchedule follows the cron format, see https://en.wikipedia.org/wiki/Cron
	CleanupSchedule string `json:"cleanupSchedule"`
	// +optional
	CleanupTimeoutSeconds *int32 `json:"cleanupTimeoutSeconds,omitempty"`
}

// QualifiedName is the cluster fully qualified name which follows the format <namespace>.<name>
func (c *Cassandra) QualifiedName() string {
	return fmt.Sprintf("%s.%s", c.Namespace, c.Name)
}

// SnapshotJobName is the name of the snapshot job for the cluster
func (c *Cassandra) SnapshotJobName() string {
	return fmt.Sprintf("%s-snapshot", c.Name)
}

// SnapshotCleanupJobName is the name of the snapshot cleanup job for the cluster
func (c *Cassandra) SnapshotCleanupJobName() string {
	return fmt.Sprintf("%s-snapshot-cleanup", c.Name)
}

// ServiceName is the cluster service name
func (c *Cassandra) ServiceName() string {
	return fmt.Sprintf("%s.%s", c.Name, c.Namespace)
}

// StorageVolumeName is the name of the volume used for storing Cassandra data
func (c *Cassandra) StorageVolumeName() string {
	return fmt.Sprintf("cassandra-storage-%s", c.Name)
}

// RackName is the fully qualifier name of the supplied rack within the cluster
func (c *Cassandra) RackName(rack *Rack) string {
	return fmt.Sprintf("%s-%s", c.Name, rack.Name)
}

// CustomConfigMapName returns the expected config map name for this cluster. This will return a value even if the config map does not exist.
func (c *Cassandra) CustomConfigMapName() string {
	return fmt.Sprintf("%s-config", c.Name)
}

// TODO: Move this to Cassandra top-level object
func (c *CassandraSpec) GetDatacenter() string {
	if c.Datacenter == nil {
		return DefaultDCName
	}
	return *c.Datacenter
}

// SnapshotPropertiesUpdated returns false when snapshot1 and snapshot2 have the same properties disregarding retention policy
func SnapshotPropertiesUpdated(snapshot1 *Snapshot, snapshot2 *Snapshot) bool {
	return snapshot1.Schedule != snapshot2.Schedule ||
		*snapshot1.TimeoutSeconds != *snapshot2.TimeoutSeconds ||
		!reflect.DeepEqual(snapshot1.Keyspaces, snapshot2.Keyspaces)
}

// SnapshotCleanupPropertiesUpdated returns false snapshot1 and snapshot2 have the same retention policy regardless of whether it is enabled or not
func SnapshotCleanupPropertiesUpdated(snapshot1 *Snapshot, snapshot2 *Snapshot) bool {
	return snapshot1.RetentionPolicy != nil && snapshot2.RetentionPolicy != nil &&
		(snapshot1.RetentionPolicy.CleanupSchedule != snapshot2.RetentionPolicy.CleanupSchedule ||
			*snapshot1.RetentionPolicy.CleanupTimeoutSeconds != *snapshot2.RetentionPolicy.CleanupTimeoutSeconds ||
			*snapshot1.RetentionPolicy.RetentionPeriodDays != *snapshot2.RetentionPolicy.RetentionPeriodDays)
}
