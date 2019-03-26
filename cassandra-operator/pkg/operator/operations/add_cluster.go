package operations

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"strings"
)

// AddClusterOperation describes what the operator does when creating a cluster
type AddClusterOperation struct {
	clusterAccessor     *cluster.Accessor
	clusters            map[string]*cluster.Cluster
	statefulSetAccessor *statefulSetAccessor
	clusterDefinition   *v1alpha1.Cassandra
}

// Execute performs the operation
func (o *AddClusterOperation) Execute() {
	log.Infof("New Cassandra cluster definition added: %s.%s", o.clusterDefinition.Namespace, o.clusterDefinition.Name)
	configMap := o.clusterAccessor.FindCustomConfigMap(o.clusterDefinition.Namespace, o.clusterDefinition.Name)
	if configMap != nil {
		log.Infof("Found custom config map for cluster %s.%s", o.clusterDefinition.Namespace, o.clusterDefinition.Name)
	}

	c, err := cluster.New(o.clusterDefinition)
	if err != nil {
		log.Errorf("Unable to create cluster %s.%s: %v", o.clusterDefinition.Namespace, o.clusterDefinition.Name, err)
		return
	}
	o.clusters[c.QualifiedName()] = c

	foundResources := o.clusterAccessor.FindExistingResourcesFor(c)
	if len(foundResources) > 0 {
		log.Infof("Resources already found for cluster %s, not attempting to recreate: %s", c.QualifiedName(), strings.Join(foundResources, ","))
	} else {
		_, err = o.clusterAccessor.CreateServiceForCluster(c)
		if err != nil {
			log.Errorf("Error while creating headless service for cluster %s: %v", c.QualifiedName(), err)
			return
		}
		log.Infof("Headless service created for cluster : %s", c.QualifiedName())

		err = o.statefulSetAccessor.registerStatefulSets(c, configMap)
		if err != nil {
			log.Errorf("Error while creating stateful sets for cluster %s: %v", c.QualifiedName(), err)
			return
		}
	}

	c.Online = true
}

func (o *AddClusterOperation) String() string {
	return fmt.Sprintf("add cluster %s", o.clusterDefinition.QualifiedName())
}
