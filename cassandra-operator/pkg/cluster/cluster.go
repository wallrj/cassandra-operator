package cluster

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron"
	appsv1 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	v1alpha1helpers "github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1/helpers"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/operator/hash"
)

const (
	// OperatorLabel is a label used on all kubernetes resources created by this Operator
	OperatorLabel = "sky.uk/cassandra-operator"

	// ConfigHashAnnotation gives the name of the annotation that the operator attaches to pods when they have
	// an associated custom config map.
	ConfigHashAnnotation = "clusterConfigHash"

	// RackLabel is a label used to identify the rack name in a cluster
	RackLabel       = "rack"
	customConfigDir = "/custom-config"

	cassandraContainerName             = "cassandra"
	cassandraBootstrapperContainerName = "cassandra-bootstrapper"
	cassandraSidecarContainerName      = "cassandra-sidecar"

	storageVolumeMountPath       = "/var/lib/cassandra"
	configurationVolumeMountPath = "/etc/cassandra"
	extraLibVolumeMountPath      = "/extra-lib"
	configurationVolumeName      = "configuration"
	extraLibVolumeName           = "extra-lib"

	healthServerPort = 8080
)

var (
	maxSidecarMemoryRequest resource.Quantity
	sidecarMemoryLimit      resource.Quantity
	maxSidecarCPURequest    resource.Quantity
	sidecarCPULimit         resource.Quantity
)

func init() {
	maxSidecarMemoryRequest = resource.MustParse("50Mi")
	sidecarMemoryLimit = resource.MustParse("50Mi")
	maxSidecarCPURequest = resource.MustParse("100m")
	sidecarCPULimit = resource.MustParse("200m")
}

// Cluster defines the properties of a Cassandra cluster which the operator should manage.
type Cluster struct {
	definition *v1alpha1.Cassandra
	Online     bool
}

// New creates a new cluster definition from the supplied Cassandra definition
func New(clusterDefinition *v1alpha1.Cassandra) (*Cluster, error) {
	cluster := &Cluster{}
	if err := CopyInto(cluster, clusterDefinition); err != nil {
		return nil, err
	}
	return cluster, nil
}

// NewWithoutValidation creates a new cluster definition from the supplied Cassandra definition without performing validation
func NewWithoutValidation(clusterDefinition *v1alpha1.Cassandra) *Cluster {
	cluster := &Cluster{
		definition: clusterDefinition,
	}
	return cluster
}

// CopyInto copies a Cassandra cluster definition into the internal cluster data structure supplied.
func CopyInto(cluster *Cluster, clusterDefinition *v1alpha1.Cassandra) error {
	if err := validateRacks(clusterDefinition); err != nil {
		return err
	}

	if err := validatePodResources(clusterDefinition); err != nil {
		return err
	}

	if err := validateSnapshot(clusterDefinition); err != nil {
		return err
	}

	if err := validateLivenessProbe(clusterDefinition.Spec.Pod.LivenessProbe, clusterDefinition); err != nil {
		return err
	}

	if err := validateReadinessProbe(clusterDefinition.Spec.Pod.ReadinessProbe, clusterDefinition); err != nil {
		return err
	}

	cluster.definition = clusterDefinition.DeepCopy()
	return nil
}

// Definition returns a copy of the definition of the cluster. Any modifications made to this will be ignored.
func (c *Cluster) Definition() *v1alpha1.Cassandra {
	return c.definition.DeepCopy()
}

func validateRacks(clusterDefinition *v1alpha1.Cassandra) error {
	if len(clusterDefinition.Spec.Racks) == 0 {
		return fmt.Errorf("no racks specified for cluster: %s.%s", clusterDefinition.Namespace, clusterDefinition.Name)
	}

	for _, rack := range clusterDefinition.Spec.Racks {
		if rack.Replicas < 1 {
			return fmt.Errorf("invalid rack replicas value %d provided for Cassandra cluster definition: %s.%s", rack.Replicas, clusterDefinition.Namespace, clusterDefinition.Name)
		} else if rack.StorageClass == "" && !*clusterDefinition.Spec.UseEmptyDir {
			return fmt.Errorf("rack named '%s' with no storage class specified, either set useEmptyDir to true or specify storage class: %s.%s", rack.Name, clusterDefinition.Namespace, clusterDefinition.Name)
		} else if rack.Zone == "" && !*clusterDefinition.Spec.UseEmptyDir {
			return fmt.Errorf("rack named '%s' with no zone specified, either set useEmptyDir to true or specify zone: %s.%s", rack.Name, clusterDefinition.Namespace, clusterDefinition.Name)
		}
	}
	return nil
}

func validatePodResources(clusterDefinition *v1alpha1.Cassandra) error {
	if clusterDefinition.Spec.Pod.Memory.IsZero() {
		return fmt.Errorf("no podMemory property provided for Cassandra cluster definition: %s.%s", clusterDefinition.Namespace, clusterDefinition.Name)
	}

	if *clusterDefinition.Spec.UseEmptyDir && !clusterDefinition.Spec.Pod.StorageSize.IsZero() {
		return fmt.Errorf("podStorageSize property provided when useEmptyDir is true for Cassandra cluster definition: %s.%s", clusterDefinition.Namespace, clusterDefinition.Name)
	}

	if !*clusterDefinition.Spec.UseEmptyDir && clusterDefinition.Spec.Pod.StorageSize.IsZero() {
		return fmt.Errorf("no podStorageSize property provided and useEmptyDir false for Cassandra cluster definition: %s.%s", clusterDefinition.Namespace, clusterDefinition.Name)
	}
	return nil
}

func validateSnapshot(clusterDefinition *v1alpha1.Cassandra) error {
	if clusterDefinition.Spec.Snapshot == nil {
		return nil
	}

	if clusterDefinition.Spec.Snapshot.Schedule == "" {
		return fmt.Errorf("no snapshot schedule property provided for Cassandra cluster definition: %s", clusterDefinition.QualifiedName())
	}

	if _, err := cron.Parse(clusterDefinition.Spec.Snapshot.Schedule); err != nil {
		return fmt.Errorf("invalid snapshot schedule, must be a cron expression but got '%s' for Cassandra cluster definition: %s.%s", clusterDefinition.Spec.Snapshot.Schedule, clusterDefinition.Namespace, clusterDefinition.Name)
	}

	timeoutSeconds := clusterDefinition.Spec.Snapshot.TimeoutSeconds
	if timeoutSeconds != nil && *timeoutSeconds < 0 {
		return fmt.Errorf("invalid snapshot timeoutSeconds value %d, must be non-negative for Cassandra cluster definition: %s", *timeoutSeconds, clusterDefinition.QualifiedName())
	}

	retentionPolicy := clusterDefinition.Spec.Snapshot.RetentionPolicy
	if retentionPolicy != nil {
		retentionPeriodDays := retentionPolicy.RetentionPeriodDays
		if retentionPeriodDays != nil && *retentionPeriodDays < 0 {
			return fmt.Errorf("invalid snapshot retention policy retentionPeriodDays value %d, must be non-negative for Cassandra cluster definition: %s", *retentionPeriodDays, clusterDefinition.QualifiedName())
		}

		cleanupTimeoutSeconds := retentionPolicy.CleanupTimeoutSeconds
		if cleanupTimeoutSeconds != nil && *cleanupTimeoutSeconds < 0 {
			return fmt.Errorf("invalid snapshot retention policy cleanupTimeoutSeconds value %d, must be non-negative for Cassandra cluster definition: %s", *cleanupTimeoutSeconds, clusterDefinition.QualifiedName())
		}

		if retentionPolicy.CleanupSchedule != "" {
			if _, err := cron.Parse(retentionPolicy.CleanupSchedule); err != nil {
				return fmt.Errorf("invalid snapshot cleanup schedule, must be a cron expression but got '%s' for Cassandra cluster definition: %s", retentionPolicy.CleanupSchedule, clusterDefinition.QualifiedName())
			}
		}
	}

	return nil
}

func validateLivenessProbe(probe *v1alpha1.Probe, clusterDefinition *v1alpha1.Cassandra) error {
	if probe.SuccessThreshold == nil || *probe.SuccessThreshold != 1 {
		return fmt.Errorf("invalid success threshold for liveness probe, must be set to 1 for Cassandra cluster definition: %s.%s", clusterDefinition.Namespace, clusterDefinition.Name)
	}
	return validateProbe("liveness", probe, clusterDefinition)
}

func validateReadinessProbe(probe *v1alpha1.Probe, clusterDefinition *v1alpha1.Cassandra) error {
	return validateProbe("readiness", probe, clusterDefinition)
}

func validateProbe(name string, probe *v1alpha1.Probe, clusterDefinition *v1alpha1.Cassandra) error {
	if probe.FailureThreshold == nil || *probe.FailureThreshold < 1 {
		return fmt.Errorf("invalid failure threshold for %s probe, must be 1 or greater, got %d for Cassandra cluster definition: %s.%s", name, *probe.FailureThreshold, clusterDefinition.Namespace, clusterDefinition.Name)
	}
	if probe.InitialDelaySeconds == nil || *probe.InitialDelaySeconds < 1 {
		return fmt.Errorf("invalid initial delay for %s probe, must be 1 or greater, got %d for Cassandra cluster definition: %s.%s", name, *probe.InitialDelaySeconds, clusterDefinition.Namespace, clusterDefinition.Name)
	}
	if probe.PeriodSeconds == nil || *probe.PeriodSeconds < 1 {
		return fmt.Errorf("invalid period seconds for %s probe, must be 1 or greater, got %d for Cassandra cluster definition: %s.%s", name, *probe.PeriodSeconds, clusterDefinition.Namespace, clusterDefinition.Name)
	}
	if probe.SuccessThreshold == nil || *probe.SuccessThreshold < 1 {
		return fmt.Errorf("invalid success threshold for %s probe, must be 1 or greater, got %d for Cassandra cluster definition: %s.%s", name, *probe.SuccessThreshold, clusterDefinition.Namespace, clusterDefinition.Name)
	}
	if probe.TimeoutSeconds == nil || *probe.TimeoutSeconds < 1 {
		return fmt.Errorf("invalid timeout seconds for %s probe, must be 1 or greater, got %d for Cassandra cluster definition: %s.%s", name, *probe.TimeoutSeconds, clusterDefinition.Namespace, clusterDefinition.Name)
	}
	return nil
}

func mergeProbeDefaults(configuredProbe *v1alpha1.Probe, defaultProbe *v1alpha1.Probe) {
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

// Name is the unqualified name of the cluster
func (c *Cluster) Name() string {
	return c.definition.Name
}

// Namespace is the namespace the cluster resides in
func (c *Cluster) Namespace() string {
	return c.definition.Namespace
}

// QualifiedName is the namespace-qualified name of the cluster
func (c *Cluster) QualifiedName() string {
	return c.definition.QualifiedName()
}

// Racks returns the set of racks defined for the cluster
func (c *Cluster) Racks() []v1alpha1.Rack {
	return c.definition.Spec.Racks
}

func (c *Cluster) createStatefulSetForRack(rack *v1alpha1.Rack, customConfigMap *v1.ConfigMap) *appsv1.StatefulSet {
	sts := &appsv1.StatefulSet{
		ObjectMeta: c.objectMetadataWithOwner(c.definition.RackName(rack), RackLabel, rack.Name),
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					OperatorLabel: c.definition.Name,
					RackLabel:     rack.Name,
					"app":         c.definition.Name,
				},
			},
			Replicas:    &rack.Replicas,
			ServiceName: c.definition.Name,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						OperatorLabel: c.definition.Name,
						RackLabel:     rack.Name,
						"app":         c.definition.Name,
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: v1alpha1.NodeServiceAccountName,
					InitContainers: []v1.Container{
						c.createInitConfigContainer(),
						c.createCassandraBootstrapperContainer(rack, customConfigMap),
					},
					Containers: []v1.Container{
						c.createCassandraContainer(rack, customConfigMap),
						c.createCassandraSidecarContainer(rack),
					},
					Volumes: c.createPodVolumes(customConfigMap),
					Affinity: &v1.Affinity{
						PodAntiAffinity: &v1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											OperatorLabel: c.definition.Name,
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
						NodeAffinity: &v1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
								NodeSelectorTerms: []v1.NodeSelectorTerm{
									{
										MatchExpressions: []v1.NodeSelectorRequirement{
											{
												Key:      "failure-domain.beta.kubernetes.io/zone",
												Operator: v1.NodeSelectorOpIn,
												Values:   []string{rack.Zone},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: c.createCassandraDataPersistentVolumeClaimForRack(rack),
		},
	}

	if customConfigMap != nil {
		sts.Spec.Template.Annotations = map[string]string{ConfigHashAnnotation: hash.ConfigMapHash(customConfigMap)}
	}

	return sts
}

// CreateService creates a headless service for the supplied cluster definition.
func (c *Cluster) CreateService() *v1.Service {
	return &v1.Service{
		ObjectMeta: c.objectMetadataWithOwner(c.definition.Name, "app", c.definition.Name),
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": c.definition.Name,
			},
			ClusterIP: v1.ClusterIPNone,
			Ports: []v1.ServicePort{
				{
					Name:       "cassandra",
					Port:       9042,
					TargetPort: intstr.FromInt(9042),
				},
				{
					Name:       "jolokia",
					Port:       7777,
					TargetPort: intstr.FromInt(7777),
				},
			},
		},
	}
}

// CreateSnapshotJob creates a cronjob to trigger the creation of a snapshot
func (c *Cluster) CreateSnapshotJob() *v1beta1.CronJob {
	if c.definition.Spec.Snapshot == nil {
		return nil
	}

	return c.createCronJob(
		c.definition.SnapshotJobName(),
		v1alpha1.SnapshotServiceAccountName,
		c.definition.Spec.Snapshot.Schedule,
		c.CreateSnapshotContainer(c.definition.Spec.Snapshot),
	)
}

// CreateSnapshotContainer creates the container used to trigger the snapshot creation
func (c *Cluster) CreateSnapshotContainer(snapshot *v1alpha1.Snapshot) *v1.Container {
	backupCommand := []string{"/cassandra-snapshot", "create",
		"-n", c.Namespace(),
		"-l", fmt.Sprintf("%s=%s,%s=%s", OperatorLabel, c.Name(), "app", c.Name()),
	}
	if snapshot.TimeoutSeconds != nil {
		timeoutDuration := durationSeconds(snapshot.TimeoutSeconds)
		backupCommand = append(backupCommand, "-t", timeoutDuration.String())
	}
	if len(snapshot.Keyspaces) > 0 {
		backupCommand = append(backupCommand, "-k")
		backupCommand = append(backupCommand, strings.Join(snapshot.Keyspaces, ","))
	}

	return &v1.Container{
		Name:    c.definition.SnapshotJobName(),
		Image:   *c.definition.Spec.Snapshot.Image,
		Command: backupCommand,
	}
}

// CreateSnapshotCleanupJob creates a cronjob to trigger the snapshot cleanup
func (c *Cluster) CreateSnapshotCleanupJob() *v1beta1.CronJob {
	if c.definition.Spec.Snapshot == nil ||
		!v1alpha1helpers.HasRetentionPolicyEnabled(c.definition.Spec.Snapshot) {
		return nil
	}

	return c.createCronJob(
		c.definition.SnapshotCleanupJobName(),
		v1alpha1.SnapshotServiceAccountName,
		c.definition.Spec.Snapshot.RetentionPolicy.CleanupSchedule,
		c.CreateSnapshotCleanupContainer(c.definition.Spec.Snapshot),
	)
}

// CreateSnapshotCleanupContainer creates the container that will execute the snapshot cleanup command
func (c *Cluster) CreateSnapshotCleanupContainer(snapshot *v1alpha1.Snapshot) *v1.Container {
	cleanupCommand := []string{"/cassandra-snapshot", "cleanup",
		"-n", c.Namespace(),
		"-l", fmt.Sprintf("%s=%s,%s=%s", OperatorLabel, c.Name(), "app", c.Name()),
	}
	if snapshot.RetentionPolicy.RetentionPeriodDays != nil {
		retentionPeriodDuration := durationDays(snapshot.RetentionPolicy.RetentionPeriodDays)
		cleanupCommand = append(cleanupCommand, "-r", retentionPeriodDuration.String())
	}
	if snapshot.RetentionPolicy.CleanupTimeoutSeconds != nil {
		cleanupTimeoutDuration := durationSeconds(snapshot.RetentionPolicy.CleanupTimeoutSeconds)
		cleanupCommand = append(cleanupCommand, "-t", cleanupTimeoutDuration.String())
	}

	return &v1.Container{
		Name:    c.definition.SnapshotCleanupJobName(),
		Image:   *c.definition.Spec.Snapshot.Image,
		Command: cleanupCommand,
	}
}

func (c *Cluster) createCronJob(objectName, serviceAccountName, schedule string, container *v1.Container) *v1beta1.CronJob {
	return &v1beta1.CronJob{
		ObjectMeta: c.objectMetadataWithOwner(objectName, "app", objectName),
		Spec: v1beta1.CronJobSpec{
			Schedule:          schedule,
			ConcurrencyPolicy: v1beta1.ForbidConcurrent,
			JobTemplate: v1beta1.JobTemplateSpec{
				ObjectMeta: c.objectMetadataWithOwner(objectName, "app", objectName),
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						ObjectMeta: c.objectMetadataWithOwner(objectName, "app", objectName),
						Spec: v1.PodSpec{
							RestartPolicy:      v1.RestartPolicyOnFailure,
							ServiceAccountName: serviceAccountName,
							Containers:         []v1.Container{*container},
						},
					},
				},
			},
		},
	}
}

func (c *Cluster) objectMetadata(name string, extraLabels ...string) metav1.ObjectMeta {
	labels := map[string]string{OperatorLabel: c.Name()}
	for i := 0; i < len(extraLabels)-1; i += 2 {
		labels[extraLabels[i]] = extraLabels[i+1]
	}

	return metav1.ObjectMeta{
		Name:      name,
		Namespace: c.Namespace(),
		Labels:    labels,
	}
}

func (c *Cluster) objectMetadataWithOwner(name string, extraLabels ...string) metav1.ObjectMeta {
	meta := c.objectMetadata(name, extraLabels...)
	meta.OwnerReferences = []metav1.OwnerReference{v1alpha1helpers.NewControllerRef(c.definition)}
	return meta
}

func (c *Cluster) createCassandraContainer(rack *v1alpha1.Rack, customConfigMap *v1.ConfigMap) v1.Container {
	return v1.Container{
		Name:  cassandraContainerName,
		Image: *c.definition.Spec.Pod.Image,
		Ports: []v1.ContainerPort{
			{
				Name:          "internode",
				Protocol:      v1.ProtocolTCP,
				ContainerPort: 7000,
			},
			{
				Name:          "jmx-exporter",
				Protocol:      v1.ProtocolTCP,
				ContainerPort: 7070,
			},
			{
				Name:          "cassandra-jmx",
				Protocol:      v1.ProtocolTCP,
				ContainerPort: 7199,
			},
			{
				Name:          "jolokia",
				Protocol:      v1.ProtocolTCP,
				ContainerPort: 7777,
			},
			{
				Name:          "client",
				Protocol:      v1.ProtocolTCP,
				ContainerPort: 9042,
			},
		},
		Resources:      c.createResourceRequirements(),
		LivenessProbe:  createHTTPProbe(c.definition.Spec.Pod.LivenessProbe, "/live", healthServerPort),
		ReadinessProbe: createHTTPProbe(c.definition.Spec.Pod.ReadinessProbe, "/ready", healthServerPort),
		Lifecycle: &v1.Lifecycle{
			PreStop: &v1.Handler{
				Exec: &v1.ExecAction{
					Command: []string{"/bin/sh", "-c", "nodetool drain"},
				},
			},
		},
		Env: []v1.EnvVar{
			{Name: "EXTRA_CLASSPATH", Value: "/extra-lib/cassandra-seed-provider.jar"},
		},
		VolumeMounts: c.createVolumeMounts(customConfigMap),
	}
}

func (c *Cluster) createCassandraSidecarContainer(rack *v1alpha1.Rack) v1.Container {
	return v1.Container{
		Name:  cassandraSidecarContainerName,
		Image: *c.definition.Spec.Pod.SidecarImage,
		Ports: []v1.ContainerPort{
			{
				Name:          "api",
				Protocol:      v1.ProtocolTCP,
				ContainerPort: healthServerPort,
			},
		},
		Env: c.createEnvironmentVariableDefinition(rack),
		Args: []string{
			"--log-level=info",
			fmt.Sprintf("--health-server-port=%d", healthServerPort),
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: minQuantity(
					c.definition.Spec.Pod.CPU,
					maxSidecarCPURequest,
				),
				v1.ResourceMemory: minQuantity(
					c.definition.Spec.Pod.Memory,
					maxSidecarMemoryRequest,
				),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    sidecarCPULimit,
				v1.ResourceMemory: sidecarMemoryLimit,
			},
		},
	}
}

func (c *Cluster) createEnvironmentVariableDefinition(rack *v1alpha1.Rack) []v1.EnvVar {
	envVariables := []v1.EnvVar{
		{
			Name:  "CLUSTER_NAMESPACE",
			Value: c.definition.Namespace,
		},
		{
			Name:  "CLUSTER_NAME",
			Value: c.definition.Name,
		},
		{
			Name:  "CLUSTER_CURRENT_RACK",
			Value: rack.Name,
		},
		{
			Name:  "CLUSTER_DATA_CENTER",
			Value: *c.definition.Spec.Datacenter,
		},
		{
			Name: "NODE_LISTEN_ADDRESS",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name:  "POD_CPU_MILLICORES",
			Value: fmt.Sprintf("%d", c.definition.Spec.Pod.CPU.MilliValue()),
		},
		{
			Name:  "POD_MEMORY_BYTES",
			Value: fmt.Sprintf("%d", c.definition.Spec.Pod.Memory.Value()),
		},
	}

	return envVariables
}

func (c *Cluster) createCassandraDataPersistentVolumeClaimForRack(rack *v1alpha1.Rack) []v1.PersistentVolumeClaim {
	var persistentVolumeClaim []v1.PersistentVolumeClaim

	if !*c.definition.Spec.UseEmptyDir {
		persistentVolumeClaim = append(persistentVolumeClaim, v1.PersistentVolumeClaim{
			ObjectMeta: c.objectMetadata(c.definition.StorageVolumeName(), RackLabel, rack.Name, "app", c.definition.Name),
			Spec: v1.PersistentVolumeClaimSpec{
				StorageClassName: &rack.StorageClass,
				AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: c.definition.Spec.Pod.StorageSize,
					},
				},
			},
		})
	}

	return persistentVolumeClaim
}

func (c *Cluster) createVolumeMounts(customConfigMap *v1.ConfigMap) []v1.VolumeMount {
	mounts := []v1.VolumeMount{
		{Name: c.definition.StorageVolumeName(), MountPath: storageVolumeMountPath},
		{Name: configurationVolumeName, MountPath: configurationVolumeMountPath},
		{Name: extraLibVolumeName, MountPath: extraLibVolumeMountPath},
	}

	return mounts
}

func (c *Cluster) createCustomConfigVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      c.customConfigMapVolumeName(),
		MountPath: customConfigDir,
	}
}

func (c *Cluster) createPodVolumes(customConfigMap *v1.ConfigMap) []v1.Volume {
	volumes := []v1.Volume{
		emptyDir("configuration"),
		emptyDir("extra-lib"),
	}

	if customConfigMap != nil {
		volumes = append(volumes, c.createConfigMapVolume(customConfigMap))
	}

	if *c.definition.Spec.UseEmptyDir {
		volumes = append(volumes, emptyDir(c.definition.StorageVolumeName()))
	}

	return volumes
}

func emptyDir(name string) v1.Volume {
	return v1.Volume{
		Name:         name,
		VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
	}
}

func (c *Cluster) createConfigMapVolume(configMap *v1.ConfigMap) v1.Volume {
	return v1.Volume{
		Name: c.customConfigMapVolumeName(),
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: configMap.Name,
				},
			},
		},
	}
}

func createProbe(probe *v1alpha1.Probe, command ...string) *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			Exec: &v1.ExecAction{
				Command: command,
			},
		},
		InitialDelaySeconds: *probe.InitialDelaySeconds,
		PeriodSeconds:       *probe.PeriodSeconds,
		TimeoutSeconds:      *probe.TimeoutSeconds,
		FailureThreshold:    *probe.FailureThreshold,
		SuccessThreshold:    *probe.SuccessThreshold,
	}
}

func createHTTPProbe(probe *v1alpha1.Probe, path string, port int) *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Port: intstr.FromInt(port),
				Path: path,
			},
		},
		InitialDelaySeconds: *probe.InitialDelaySeconds,
		PeriodSeconds:       *probe.PeriodSeconds,
		TimeoutSeconds:      *probe.TimeoutSeconds,
		FailureThreshold:    *probe.FailureThreshold,
		SuccessThreshold:    *probe.SuccessThreshold,
	}
}

// AddCustomConfigVolumeToStatefulSet updates the provided statefulset to mount the configmap as a volume
func (c *Cluster) AddCustomConfigVolumeToStatefulSet(statefulSet *appsv1.StatefulSet, customConfigMap *v1.ConfigMap) error {
	if statefulSet.Spec.Template.Annotations == nil {
		statefulSet.Spec.Template.Annotations = map[string]string{}
	}
	statefulSet.Spec.Template.Annotations[ConfigHashAnnotation] = hash.ConfigMapHash(customConfigMap)

	statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes, c.createConfigMapVolume(customConfigMap))
	for i := range statefulSet.Spec.Template.Spec.InitContainers {
		if statefulSet.Spec.Template.Spec.InitContainers[i].Name == cassandraBootstrapperContainerName {
			statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts = append(statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts, c.createCustomConfigVolumeMount())
		}
	}
	return nil
}

// RemoveCustomConfigVolumeFromStatefulSet updates the provided statefulset to unmount the configmap as a volume
func (c *Cluster) RemoveCustomConfigVolumeFromStatefulSet(statefulSet *appsv1.StatefulSet, _ *v1.ConfigMap) error {
	var volumesAfterRemoval []v1.Volume
	for _, volume := range statefulSet.Spec.Template.Spec.Volumes {
		if volume.Name != c.customConfigMapVolumeName() {
			volumesAfterRemoval = append(volumesAfterRemoval, volume)
		}
	}
	statefulSet.Spec.Template.Spec.Volumes = volumesAfterRemoval

	delete(statefulSet.Spec.Template.Annotations, ConfigHashAnnotation)

	var volumesMountAfterRemoval []v1.VolumeMount
	for i := range statefulSet.Spec.Template.Spec.InitContainers {
		if statefulSet.Spec.Template.Spec.InitContainers[i].Name == cassandraBootstrapperContainerName {
			for _, volumeMount := range statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts {
				if volumeMount.Name != c.customConfigMapVolumeName() {
					volumesMountAfterRemoval = append(volumesMountAfterRemoval, volumeMount)
				}
			}
			statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts = volumesMountAfterRemoval
		}
	}
	return nil
}

func (c *Cluster) customConfigMapVolumeName() string {
	return fmt.Sprintf("cassandra-custom-config-%s", c.definition.Name)
}
func (c *Cluster) createInitConfigContainer() v1.Container {
	return v1.Container{
		Name:    "init-config",
		Image:   *c.definition.Spec.Pod.Image,
		Command: []string{"sh", "-c", "cp -vr /etc/cassandra/* /configuration"},
		VolumeMounts: []v1.VolumeMount{
			{Name: "configuration", MountPath: "/configuration"},
		},
		Resources: c.createResourceRequirements(),
	}
}
func (c *Cluster) createCassandraBootstrapperContainer(rack *v1alpha1.Rack, customConfigMap *v1.ConfigMap) v1.Container {
	mounts := []v1.VolumeMount{
		{Name: "configuration", MountPath: "/configuration"},
		{Name: "extra-lib", MountPath: "/extra-lib"},
	}

	if customConfigMap != nil {
		mounts = append(mounts, c.createCustomConfigVolumeMount())
	}

	return v1.Container{
		Name:         cassandraBootstrapperContainerName,
		Env:          c.createEnvironmentVariableDefinition(rack),
		Image:        *c.definition.Spec.Pod.BootstrapperImage,
		Resources:    c.createResourceRequirements(),
		VolumeMounts: mounts,
	}
}

func (c *Cluster) createResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    c.definition.Spec.Pod.CPU,
			v1.ResourceMemory: c.definition.Spec.Pod.Memory,
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory: c.definition.Spec.Pod.Memory,
		},
	}
}

// ConfigMapBelongsToAManagedCluster determines whether the supplied ConfigMap belongs to a managed cluster
func ConfigMapBelongsToAManagedCluster(managedClusters map[string]*Cluster, configMap *v1.ConfigMap) bool {
	for _, mc := range managedClusters {
		if configMap.Name == mc.definition.CustomConfigMapName() {
			return true
		}
	}
	return false
}

// LooksLikeACassandraConfigMap determines whether the supplied ConfigMap could belong to a managed cluster
func LooksLikeACassandraConfigMap(configMap *v1.ConfigMap) bool {
	return strings.HasSuffix(configMap.Name, "-config")
}

// QualifiedClusterNameFor returns the fully qualified name of the cluster that should be associated to the supplied configMap
func QualifiedClusterNameFor(configMap *v1.ConfigMap) (string, error) {
	if !LooksLikeACassandraConfigMap(configMap) {
		return "", fmt.Errorf("configMap name %s does not follow the naming convention for a cluster", configMap.Name)
	}
	return fmt.Sprintf("%s.%s", configMap.Namespace, strings.Replace(configMap.Name, "-config", "", -1)), nil
}

func durationDays(days *int32) time.Duration {
	return time.Duration(*days) * time.Hour * 24
}

func durationSeconds(seconds *int32) time.Duration {
	return time.Duration(*seconds) * time.Second
}

func minQuantity(r1, r2 resource.Quantity) resource.Quantity {
	d := r1.Cmp(r2)
	if d > 0 {
		return r2
	}
	return r1
}
