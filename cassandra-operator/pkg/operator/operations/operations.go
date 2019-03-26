package operations

import (
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"k8s.io/api/core/v1"
)

// Operation describes a single unit of work
type Operation interface {
	// Execute actually performs the operation
	Execute()
	// Human-readable description of the operation
	String() string
}

func (r *Receiver) newAddCluster(cassandra *v1alpha1.Cassandra) Operation {
	return &AddClusterOperation{
		clusterAccessor:     r.clusterAccessor,
		clusters:            r.clusters,
		statefulSetAccessor: r.statefulSetAccessor,
		clusterDefinition:   cassandra,
	}
}

func (r *Receiver) newDeleteCluster(cassandra *v1alpha1.Cassandra) Operation {
	return &DeleteClusterOperation{
		clusterAccessor:   r.clusterAccessor,
		clusters:          r.clusters,
		clusterDefinition: cassandra,
		metricsPoller:     r.metricsPoller,
	}
}

func (r *Receiver) newDeleteSnapshot(cassandra *v1alpha1.Cassandra) Operation {
	return &DeleteSnapshotOperation{
		cassandra:       cassandra,
		clusterAccessor: r.clusterAccessor,
		eventRecorder:   r.eventRecorder,
	}
}

func (r *Receiver) newDeleteSnapshotCleanup(cassandra *v1alpha1.Cassandra) Operation {
	return &DeleteSnapshotCleanupOperation{
		cassandra:       cassandra,
		clusterAccessor: r.clusterAccessor,
		eventRecorder:   r.eventRecorder,
	}
}

func (r *Receiver) newAddSnapshot(cassandra *v1alpha1.Cassandra) Operation {
	return &AddSnapshotOperation{
		clusterDefinition: cassandra,
		clusterAccessor:   r.clusterAccessor,
		eventRecorder:     r.eventRecorder,
	}
}

func (r *Receiver) newAddSnapshotCleanup(cassandra *v1alpha1.Cassandra) Operation {
	return &AddSnapshotCleanupOperation{
		clusterDefinition: cassandra,
		clusterAccessor:   r.clusterAccessor,
		eventRecorder:     r.eventRecorder,
	}
}

func (r *Receiver) newUpdateCluster(c *cluster.Cluster, update ClusterUpdate) Operation {
	return &UpdateClusterOperation{
		cluster:             c,
		adjuster:            r.adjuster,
		eventRecorder:       r.eventRecorder,
		statefulSetAccessor: r.statefulSetAccessor,
		clusterAccessor:     r.clusterAccessor,
		update:              update,
	}
}

func (r *Receiver) newUpdateSnapshot(c *cluster.Cluster, newSnapshot *v1alpha1.Snapshot) Operation {
	return &UpdateSnapshotOperation{
		cluster:         c,
		newSnapshot:     newSnapshot,
		clusterAccessor: r.clusterAccessor,
		eventRecorder:   r.eventRecorder,
	}
}

func (r *Receiver) newUpdateSnapshotCleanup(c *cluster.Cluster, newSnapshot *v1alpha1.Snapshot) Operation {
	return &UpdateSnapshotCleanupOperation{
		cluster:         c,
		newSnapshot:     newSnapshot,
		clusterAccessor: r.clusterAccessor,
		eventRecorder:   r.eventRecorder,
	}
}

func (r *Receiver) newGatherMetrics(c *cluster.Cluster) Operation {
	return &GatherMetricsOperation{metricsPoller: r.metricsPoller, cluster: c}
}

func (r *Receiver) newUpdateCustomConfig(cluster *cluster.Cluster, configMap *v1.ConfigMap) Operation {
	return &UpdateCustomConfigOperation{
		cluster:             cluster,
		configMap:           configMap,
		eventRecorder:       r.eventRecorder,
		statefulSetAccessor: r.statefulSetAccessor,
		adjuster:            r.adjuster,
	}
}

func (r *Receiver) newAddCustomConfig(cluster *cluster.Cluster, configMap *v1.ConfigMap) Operation {
	return &AddCustomConfigOperation{
		cluster:             cluster,
		configMap:           configMap,
		eventRecorder:       r.eventRecorder,
		statefulSetAccessor: r.statefulSetAccessor,
		adjuster:            r.adjuster,
	}
}

func (r *Receiver) newDeleteCustomConfig(cluster *cluster.Cluster, configMap *v1.ConfigMap) Operation {
	return &DeleteCustomConfigOperation{
		cluster:             cluster,
		configMap:           configMap,
		eventRecorder:       r.eventRecorder,
		statefulSetAccessor: r.statefulSetAccessor,
	}
}
