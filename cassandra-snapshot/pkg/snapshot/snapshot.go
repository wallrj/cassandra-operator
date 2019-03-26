package snapshot

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/cassandra-operator/cassandra-snapshot/pkg/nodetool"
	"github.com/sky-uk/cassandra-operator/cassandra-snapshot/pkg/snapshot/filter"
	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"time"
)

// CreateConfig is the configuration for backup operations
type CreateConfig struct {
	Keyspaces       []string
	Namespace       string
	PodLabel        string
	SnapshotTimeout time.Duration
}

// CleanupConfig is the configuration for backup removal operations
type CleanupConfig struct {
	Namespace       string
	RetentionPeriod time.Duration
	Keyspaces       []string
	PodLabel        string
	CleanupTimeout  time.Duration
}

// Manipulator is responsible for creating and deleting snapshots
type Manipulator struct {
	kubeconfig     *rest.Config
	kubeClient     *kubernetes.Clientset
	nodetoolClient *nodetool.Nodetool
}

// New creates a new Manipulator
func New() *Manipulator {
	kubeconfig := kubernetesConfig()
	kubeClient := kubernetesClient(kubeconfig)
	return &Manipulator{
		kubeconfig:     kubeconfig,
		kubeClient:     kubeClient,
		nodetoolClient: nodetool.New(kubeClient, kubeconfig),
	}
}

// DoCreate creates snapshots for one or more keyspaces of a cluster
func (m *Manipulator) DoCreate(config *CreateConfig) error {
	podList, err := m.kubeClient.CoreV1().Pods(config.Namespace).List(metaV1.ListOptions{LabelSelector: config.PodLabel})
	if err != nil {
		return fmt.Errorf("unable to find cassandra pods with label %s: %v", config.PodLabel, err)
	}

	if len(podList.Items) == 0 {
		return fmt.Errorf("no cassandra pods found with label %s in namespace %s", config.PodLabel, config.Namespace)
	}

	var failedPods []string
	snapshotTimestamp := time.Now()
	for _, pod := range podList.Items {
		if err := m.nodetoolClient.CreateSnapshot(snapshotTimestamp, config.Keyspaces, &pod, config.SnapshotTimeout); err != nil {
			log.Errorf("Error while taking snapshot for pod %s.%s: %v", pod.Namespace, pod.Name, err)
			failedPods = append(failedPods, pod.Name)
		}
	}

	if len(failedPods) > 0 {
		return fmt.Errorf("snapshot creation failed for pods: %v", failedPods)
	}
	return nil
}

// DoCleanup cleans up snapshots which lie outside of the retention period
func (m *Manipulator) DoCleanup(config *CleanupConfig) error {
	podList, err := m.kubeClient.CoreV1().Pods(config.Namespace).List(metaV1.ListOptions{LabelSelector: config.PodLabel})
	if err != nil {
		return fmt.Errorf("unable to find cassandra pods with label %s: %v", config.PodLabel, err)
	}

	if len(podList.Items) == 0 {
		return fmt.Errorf("no cassandra pods found with label %s in namespace %s", config.PodLabel, config.Namespace)
	}

	var failedPods []string
	retentionCutoff := time.Now().Unix() - int64(config.RetentionPeriod.Seconds())
	for _, pod := range podList.Items {
		snapshotsToDelete, err := m.findSnapshotsToDelete(&pod, config, retentionCutoff)
		if err != nil {
			log.Errorf("Unable to find snapshots to delete for pod %s.%s: %v", pod.Namespace, pod.Name, err)
			failedPods = append(failedPods, pod.Name)
		}

		for _, snapshotToDelete := range snapshotsToDelete {
			log.Infof("Triggering deletion of snapshot %v in pod %s.%s", snapshotToDelete, pod.Namespace, pod.Name)
			err = m.nodetoolClient.DeleteSnapshot(&pod, &snapshotToDelete, config.CleanupTimeout)
			if err != nil {
				log.Errorf("Error while deleting snapshot %v for pod %s.%s: %v", snapshotToDelete, pod.Namespace, pod.Name, err)
				failedPods = append(failedPods, pod.Name)
			}
		}
	}

	if len(failedPods) > 0 {
		return fmt.Errorf("snapshot cleanup failed for pods: %v", failedPods)
	}
	return nil
}

func (m *Manipulator) findSnapshotsToDelete(pod *v1.Pod, config *CleanupConfig, retentionCutoff int64) ([]nodetool.Snapshot, error) {
	snapshots, err := m.nodetoolClient.GetSnapshots(pod, config.CleanupTimeout, filter.OutsideRetentionPeriod(pod, retentionCutoff))
	if err != nil {
		return nil, err
	}

	return snapshots, nil
}

func kubernetesConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Unable to obtain in-cluster config: %v", err)
	}
	return config
}

func kubernetesClient(config *rest.Config) *kubernetes.Clientset {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to obtain clientset: %v", err)
	}
	return clientset
}
