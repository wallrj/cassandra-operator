package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// required for dep management
	"k8s.io/apimachinery/pkg/api/resource"
	_ "k8s.io/code-generator/cmd/client-gen/types"
)

const (
	NodeServiceAccountName     = "cassandra-node"
	SnapshotServiceAccountName = "cassandra-snapshot"

	// DefaultDatacenterName is the default data center name which each Cassandra pod belongs to
	DefaultDatacenterName = "dc1"

	// DefaultCassandraImage is the name of the default Docker image used on Cassandra pods
	DefaultCassandraImage = "cassandra:3.11"

	// DefaultCassandraBootstrapperImage is the name of the Docker image used to prepare the configuration for the Cassandra node before it can be started
	DefaultCassandraBootstrapperImage = "skyuk/cassandra-bootstrapper:latest"

	// DefaultCassandraSnapshotImage is the name of the Docker image used to make and cleanup snapshots
	DefaultCassandraSnapshotImage = "skyuk/cassandra-snapshot:latest"

	// DefaultCassandraSidecarImage is the name of the Docker image used to inform liveness/readiness probes
	DefaultCassandraSidecarImage = "skyuk/cassandra-sidecar:latest"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cassandra defines a Cassandra cluster
type Cassandra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraSpec   `json:"spec,omitempty"`
	Status CassandraStatus `json:"status,omitempty"`
}

// CassandraSpec is the specification for the Cassandra resource
type CassandraSpec struct {
	// +optional
	Datacenter *string `json:"datacenter,omitempty"`
	Racks      []Rack  `json:"racks"`
	// +optional
	UseEmptyDir *bool `json:"useEmptyDir,omitempty"`
	Pod         Pod   `json:"pod"`
	// +optional
	Snapshot *Snapshot `json:"snapshot,omitempty"`
}

type Probe struct {
	// +optional
	FailureThreshold *int32 `json:"failureThreshold,omitempty"`
	// +optional
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty"`
	// +optional
	PeriodSeconds *int32 `json:"periodSeconds,omitempty"`
	// +optional
	SuccessThreshold *int32 `json:"successThreshold,omitempty"`
	// +optional
	TimeoutSeconds *int32 `json:"timeoutSeconds,omitempty"`
}

type Pod struct {
	// +optional
	BootstrapperImage *string `json:"bootstrapperImage,omitempty"`

	// +optional
	SidecarImage *string `json:"sidecarImage,omitempty"`

	// +optional
	Image       *string           `json:"image,omitempty"`
	StorageSize resource.Quantity `json:"storageSize"`
	Memory      resource.Quantity `json:"memory"`
	CPU         resource.Quantity `json:"cpu"`
	// +optional
	LivenessProbe *Probe `json:"livenessProbe,omitempty"`
	// +optional
	ReadinessProbe *Probe `json:"readinessProbe,omitempty"`
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
	Image *string `json:"image,omitempty"`
	// Schedule follows the cron format, see https://en.wikipedia.org/wiki/Cron
	Schedule string `json:"schedule"`
	// +optional
	Keyspaces []string `json:"keyspaces,omitempty"`
	// +optional
	TimeoutSeconds *int32 `json:"timeoutSeconds,omitempty"`
	// +optional
	RetentionPolicy *RetentionPolicy `json:"retentionPolicy,omitempty"`
}

// RetentionPolicy defines how long the snapshots should be kept for and how often the cleanup task should run
type RetentionPolicy struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
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
