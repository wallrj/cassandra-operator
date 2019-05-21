package adjuster

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"
	"time"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	v1alpha1helpers "github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1/helpers"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/operator/hash"
	"k8s.io/api/core/v1"
)

const statefulSetPatchTemplate = `{
  "spec": {
    "replicas": {{ .Replicas }},
    "template": {
      "spec": {
		"initContainers": [{
           "name": "cassandra-bootstrapper",
           "image": "{{ .PodBootstrapperImage }}"
		}],
        "containers": [{
           "name": "cassandra",
           "livenessProbe": {
             "failureThreshold": {{ .PodLivenessProbe.FailureThreshold }},
             "initialDelaySeconds": {{ .PodLivenessProbe.InitialDelaySeconds }},
             "periodSeconds": {{ .PodLivenessProbe.PeriodSeconds }},
             "successThreshold": {{ .PodLivenessProbe.SuccessThreshold }},
             "timeoutSeconds": {{ .PodLivenessProbe.TimeoutSeconds }}
           },
           "readinessProbe": {
             "failureThreshold": {{ .PodReadinessProbe.FailureThreshold }},
             "initialDelaySeconds": {{ .PodReadinessProbe.InitialDelaySeconds }},
             "periodSeconds": {{ .PodReadinessProbe.PeriodSeconds }},
             "successThreshold": {{ .PodReadinessProbe.SuccessThreshold }},
             "timeoutSeconds": {{ .PodReadinessProbe.TimeoutSeconds }}
           },
           "resources": {
             "requests": {
               "cpu": "{{ .PodCPU }}",
               "memory": "{{ .PodMemory }}"
             },
			 "limits": {
               "memory": "{{ .PodMemory }}"
             }
	       }
        }]
      }
    }
  }
}`

const updateAnnotationPatchFormat = `{
  "spec": {
    "template": {
      "metadata": {
		"annotations": {
			"%s": "%s"
		}
      }
    }
  }
}`

const scaleDownPatchTemplate = `{"spec": {"replicas": %d}}`

// ClusterChangeType describes the type of change which needs to be made to a cluster.
type ClusterChangeType string

const (
	deleteRack ClusterChangeType = "delete rack"
	// AddRack means that a new rack should be added to a cluster.
	AddRack ClusterChangeType = "add rack"
	// UpdateRack means that an existing rack in the cluster needs to be updated.
	UpdateRack    ClusterChangeType = "update rack"
	scaleDownRack ClusterChangeType = "scale down rack"
)

// ClusterChange describes a single change which needs to be applied to Kubernetes in order for the running cluster to
// match the requirements described in the cluster definition.
type ClusterChange struct {
	// This is not a pointer on purpose to isolate the change from the actual state
	Rack             v1alpha1.Rack
	ChangeType       ClusterChangeType
	Patch            string
	nodesToScaleDown int
}

// Adjuster calculates the set of changes which need to be applied to Kubernetes in order for the running
// cluster to match the requirements described in the cluster definition.
type Adjuster struct {
	patchTemplate *template.Template
}

type patchProperties struct {
	Replicas             int32
	PodBootstrapperImage string
	PodCPU               string
	PodMemory            string
	PodLivenessProbe     *v1alpha1.Probe
	PodReadinessProbe    *v1alpha1.Probe
}

// New creates a new Adjuster.
func New() (*Adjuster, error) {
	tmpl := template.New("cassandra-spec-patch")
	tmpl, err := tmpl.Parse(statefulSetPatchTemplate)
	if err != nil {
		return nil, fmt.Errorf("unable to parse cassandra patch template %s: %v", statefulSetPatchTemplate, err)
	}
	return &Adjuster{tmpl}, nil
}

// ChangesForCluster compares oldCluster with newCluster, and produces an ordered list of ClusterChanges which need to
// be applied in order for the running cluster to be in the state matching newCluster.
func (r *Adjuster) ChangesForCluster(oldCluster *v1alpha1.Cassandra, newCluster *v1alpha1.Cassandra) ([]ClusterChange, error) {
	addedRacks, matchedRacks, deletedRacks := r.matchRacks(&oldCluster.Spec, &newCluster.Spec)
	if err := r.ensureChangeIsAllowed(oldCluster, newCluster, matchedRacks); err != nil {
		return nil, err
	}

	changeTime := time.Now()
	var clusterChanges []ClusterChange

	for _, addedRack := range addedRacks {
		clusterChanges = append(clusterChanges, ClusterChange{Rack: addedRack, ChangeType: AddRack})
	}

	for _, deletedRack := range deletedRacks {
		clusterChanges = append(clusterChanges, ClusterChange{Rack: deletedRack, ChangeType: deleteRack})
	}

	scaledDownRacks := r.scaledDownRacks(matchedRacks)
	for _, scaledDownRack := range scaledDownRacks {
		nodesToScaleDown := scaledDownRack.old.Replicas - scaledDownRack.new.Replicas
		clusterChanges = append(clusterChanges, ClusterChange{Rack: scaledDownRack.new, ChangeType: scaleDownRack, Patch: r.scaleDownPatchForRack(int(nodesToScaleDown)), nodesToScaleDown: int(nodesToScaleDown)})
	}

	if r.podSpecHasChanged(oldCluster, newCluster) {
		for _, matchedRack := range matchedRacks {
			clusterChanges = append(clusterChanges, ClusterChange{Rack: matchedRack.new, ChangeType: UpdateRack, Patch: r.patchForRack(&matchedRack.new, newCluster, changeTime)})
		}
	} else {
		for _, matchedRack := range r.scaledUpRacks(matchedRacks) {
			clusterChanges = append(clusterChanges, ClusterChange{Rack: matchedRack, ChangeType: UpdateRack, Patch: r.patchForRack(&matchedRack, newCluster, changeTime)})
		}
	}

	return clusterChanges, nil
}

// CreateConfigMapHashPatchForRack produces a ClusterChange which need to be applied for the given rack
func (r *Adjuster) CreateConfigMapHashPatchForRack(rack *v1alpha1.Rack, configMap *v1.ConfigMap) *ClusterChange {
	configMapHash := hash.ConfigMapHash(configMap)
	patch := fmt.Sprintf(updateAnnotationPatchFormat, cluster.ConfigHashAnnotation, configMapHash)
	return &ClusterChange{Rack: *rack, ChangeType: UpdateRack, Patch: patch}
}

func (r *Adjuster) patchForRack(rack *v1alpha1.Rack, newCluster *v1alpha1.Cassandra, changeTime time.Time) string {
	props := patchProperties{
		Replicas:             rack.Replicas,
		PodBootstrapperImage: v1alpha1helpers.GetBootstrapperImage(newCluster),
		PodCPU:               newCluster.Spec.Pod.CPU.String(),
		PodMemory:            newCluster.Spec.Pod.Memory.String(),
		PodLivenessProbe:     newCluster.Spec.Pod.LivenessProbe,
		PodReadinessProbe:    newCluster.Spec.Pod.ReadinessProbe,
	}
	var patch bytes.Buffer
	r.patchTemplate.Execute(&patch, props)
	patchString := patch.String()
	return patchString
}

func (r *Adjuster) scaleDownPatchForRack(nodesToScaleDown int) string {
	return fmt.Sprintf(scaleDownPatchTemplate, nodesToScaleDown)
}

func (r *Adjuster) ensureChangeIsAllowed(oldCluster, newCluster *v1alpha1.Cassandra, matchedRacks []matchedRack) error {
	if v1alpha1helpers.GetDatacenter(oldCluster) != v1alpha1helpers.GetDatacenter(newCluster) {
		return fmt.Errorf("changing dc is forbidden. The dc used will continue to be '%v'", v1alpha1helpers.GetDatacenter(oldCluster))
	}

	if !reflect.DeepEqual(oldCluster.Spec.Pod.Image, newCluster.Spec.Pod.Image) {
		currentImage := v1alpha1helpers.GetCassandraImage(oldCluster)
		return fmt.Errorf("changing image is forbidden. The image used will continue to be '%v'", currentImage)
	}
	if !reflect.DeepEqual(oldCluster.Spec.UseEmptyDir, newCluster.Spec.UseEmptyDir) {
		return fmt.Errorf("changing useEmptyDir is forbidden. The useEmptyDir used will continue to be '%v'", v1alpha1helpers.UseEmptyDir(oldCluster))
	}

	for _, matchedRack := range matchedRacks {
		if matchedRack.new.StorageClass != matchedRack.old.StorageClass {
			return fmt.Errorf("changing storageClass for rack '%s' is forbidden. The storageClass used will continue to be '%s'", matchedRack.old.Name, matchedRack.old.StorageClass)
		}

		if matchedRack.new.Zone != matchedRack.old.Zone {
			return fmt.Errorf("changing zone for rack '%s' is forbidden. The zone used will continue to be '%s'", matchedRack.old.Name, matchedRack.old.Zone)
		}
	}
	return nil
}

func (r *Adjuster) podSpecHasChanged(oldCluster, newCluster *v1alpha1.Cassandra) bool {
	return !reflect.DeepEqual(oldCluster.Spec.Pod.CPU, newCluster.Spec.Pod.CPU) ||
		!reflect.DeepEqual(oldCluster.Spec.Pod.Memory, newCluster.Spec.Pod.Memory) ||
		!reflect.DeepEqual(oldCluster.Spec.Pod.LivenessProbe, newCluster.Spec.Pod.LivenessProbe) ||
		!reflect.DeepEqual(oldCluster.Spec.Pod.ReadinessProbe, newCluster.Spec.Pod.ReadinessProbe) ||
		!reflect.DeepEqual(oldCluster.Spec.Pod.BootstrapperImage, newCluster.Spec.Pod.BootstrapperImage)
}

func (r *Adjuster) scaledUpRacks(matchedRacks []matchedRack) []v1alpha1.Rack {
	var scaledUpRacks []v1alpha1.Rack
	for _, matchedRack := range matchedRacks {
		if matchedRack.new.Replicas > matchedRack.old.Replicas {
			scaledUpRacks = append(scaledUpRacks, matchedRack.new)
		}
	}
	return scaledUpRacks
}

func (r *Adjuster) scaledDownRacks(matchedRacks []matchedRack) []matchedRack {
	var scaledDownRacks []matchedRack
	for _, matchedRack := range matchedRacks {
		if matchedRack.new.Replicas < matchedRack.old.Replicas {
			scaledDownRacks = append(scaledDownRacks, matchedRack)
		}
	}
	return scaledDownRacks
}

type matchedRack struct {
	old v1alpha1.Rack
	new v1alpha1.Rack
}

func (r *Adjuster) matchRacks(oldCluster, newCluster *v1alpha1.CassandraSpec) ([]v1alpha1.Rack, []matchedRack, []v1alpha1.Rack) {
	var removedRacks []v1alpha1.Rack
	var matchedRacks []matchedRack

	for _, oldRack := range oldCluster.Racks {
		if foundRack, ok := findRack(oldRack, newCluster.Racks); ok {
			matchedRacks = append(matchedRacks, matchedRack{old: oldRack, new: *foundRack})
		} else {
			removedRacks = append(removedRacks, oldRack)
		}
	}

	var addedRacks []v1alpha1.Rack
	for _, newClusterRack := range newCluster.Racks {
		if _, ok := findRack(newClusterRack, oldCluster.Racks); !ok {
			addedRacks = append(addedRacks, newClusterRack)
		}
	}
	return addedRacks, matchedRacks, removedRacks
}

func findRack(rackToFind v1alpha1.Rack, racks []v1alpha1.Rack) (*v1alpha1.Rack, bool) {
	for _, rack := range racks {
		if rack.Name == rackToFind.Name {
			return &rack, true
		}
	}
	return nil, false
}
