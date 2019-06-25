package operator

import (
	"time"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/operator/operations"

	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	v1alpha1helpers "github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1/helpers"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/client/clientset/versioned"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/dispatcher"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/metrics"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/util/ptr"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// The Operator itself.
type Operator struct {
	clusters           map[string]*cluster.Cluster
	kubeClientset      *kubernetes.Clientset
	cassandraClientset *versioned.Clientset
	dynamicClient      dynamic.Interface
	metricsPoller      *metrics.PrometheusMetrics
	config             *Config
	eventDispatcher    dispatcher.Dispatcher
	stopCh             chan struct{}
}

// The Config for the Operator
type Config struct {
	MetricPollInterval    time.Duration
	MetricRequestDuration time.Duration
	AllowEmptyDir         bool
}

const resourceResyncInterval = 5 * time.Minute

// New creates a new Operator.
func New(kubeClientset *kubernetes.Clientset, cassandraClientset *versioned.Clientset, dynamicClient dynamic.Interface, operatorConfig *Config) *Operator {
	clusters := make(map[string]*cluster.Cluster)
	metricsPoller := metrics.NewMetrics(kubeClientset.CoreV1(), &metrics.Config{RequestTimeout: operatorConfig.MetricRequestDuration})

	eventRecorder := cluster.NewEventRecorder(kubeClientset)
	clusterAccessor := cluster.NewAccessor(kubeClientset, cassandraClientset, eventRecorder)
	receiver := operations.NewEventReceiver(
		clusters,
		clusterAccessor,
		metricsPoller,
		eventRecorder,
	)

	stopCh := make(chan struct{})

	return &Operator{
		kubeClientset:      kubeClientset,
		dynamicClient:      dynamicClient,
		cassandraClientset: cassandraClientset,
		config:             operatorConfig,
		clusters:           clusters,
		eventDispatcher:    dispatcher.New(receiver.Receive, stopCh),
		stopCh:             stopCh,
		metricsPoller:      metricsPoller,
	}
}

// Run starts the Operator.
func (o *Operator) Run() {
	ns := os.Getenv("OPERATOR_NAMESPACE")
	if ns == "" {
		log.Info("Operator listening for changes in any namespace")
	} else {
		log.Infof("Operator listening for changes in namespace %s", ns)
	}

	cassandraInformer := registerCassandraInformer(o, ns)
	configMapInformer := registerConfigMapInformer(o, ns)

	o.startServer(o.metricsPoller)
	o.addSignalHandler(o.stopCh)

	// this should be replaced with cassandraInformer.Start() once we have
	// https://github.com/kubernetes/kubernetes/pull/77945 in a release
	go cassandraInformer.Informer().Run(o.stopCh)
	configMapInformer.Run(o.stopCh)
	<-o.stopCh
	log.Info("Operator shutting down")
}

func registerCassandraInformer(o *Operator, ns string) informers.GenericInformer {
	resource := schema.GroupVersionResource{Group: cassandra.GroupName, Version: cassandra.Version, Resource: cassandra.Plural}

	// this should be replaced with NewFilteredDynamicSharedInformerFactory
	// once we have https://github.com/kubernetes/kubernetes/pull/77945 in a
	// release
	cassandraInformer := dynamicinformer.NewFilteredDynamicInformer(o.dynamicClient, resource, ns, resourceResyncInterval, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, nil)
	cassandraInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.clusterAdded,
		UpdateFunc: o.clusterUpdated,
		DeleteFunc: o.clusterDeleted,
	})

	return cassandraInformer
}

func registerConfigMapInformer(o *Operator, ns string) cache.Controller {
	listWatch := cache.NewListWatchFromClient(o.kubeClientset.CoreV1().RESTClient(), "configmaps", ns, fields.Everything())
	_, informer := cache.NewInformer(listWatch, &v1.ConfigMap{}, resourceResyncInterval, cache.ResourceEventHandlerFuncs{
		AddFunc:    o.configMapAdded,
		UpdateFunc: o.configMapUpdated,
		DeleteFunc: o.configMapDeleted,
	})
	return informer
}

func (o *Operator) configMapAdded(obj interface{}) {
	cm := o.safeGetConfigMap(obj)

	if cluster.LooksLikeACassandraConfigMap(cm) {
		clusterID := fmt.Sprintf("%s.%s", cm.Namespace, cm.Name)
		o.eventDispatcher.Dispatch(&dispatcher.Event{Kind: operations.AddCustomConfig, Key: clusterID, Data: cm})
	}
}

func (o *Operator) configMapDeleted(obj interface{}) {
	cm := o.safeGetConfigMap(obj)

	if cluster.ConfigMapBelongsToAManagedCluster(o.clusters, cm) {
		clusterID := fmt.Sprintf("%s.%s", cm.Namespace, cm.Name)
		o.eventDispatcher.Dispatch(&dispatcher.Event{Kind: operations.DeleteCustomConfig, Key: clusterID, Data: cm})
	}
}

func (o *Operator) configMapUpdated(old interface{}, new interface{}) {
	oldConfigMap := o.safeGetConfigMap(old)
	newConfigMap := o.safeGetConfigMap(new)

	if reflect.DeepEqual(oldConfigMap.Data, newConfigMap.Data) {
		log.Debugf("update event received for config map %s.%s but no changes detected", newConfigMap.Namespace, newConfigMap.Name)
		return
	}

	if cluster.ConfigMapBelongsToAManagedCluster(o.clusters, oldConfigMap) {
		clusterID := fmt.Sprintf("%s.%s", oldConfigMap.Namespace, oldConfigMap.Name)
		o.eventDispatcher.Dispatch(&dispatcher.Event{Kind: operations.UpdateCustomConfig, Key: clusterID, Data: newConfigMap})
	}
}

// DeepCopy the object we get back from the informer to avoid modified the "cached" object
func (o *Operator) safeGetConfigMap(obj interface{}) *v1.ConfigMap {
	cm := obj.(*v1.ConfigMap)
	return cm.DeepCopy()
}

func (o *Operator) clusterAdded(obj interface{}) {
	clusterDefinition, err := unstructuredToCassandra(obj)
	if err != nil {
		logCassandraDecodeError(obj, err)
		return
	}

	v1alpha1helpers.SetDefaultsForCassandra(clusterDefinition)
	o.adjustUseEmptyDir(clusterDefinition)

	clusterID := fmt.Sprintf("%s.%s", clusterDefinition.Namespace, clusterDefinition.Name)
	o.eventDispatcher.Dispatch(&dispatcher.Event{Kind: operations.AddCluster, Key: clusterID, Data: clusterDefinition})
}

func unstructuredToCassandra(obj interface{}) (*v1alpha1.Cassandra, error) {
	un, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("object is not an unstructured")
	}

	var c v1alpha1.Cassandra
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func logCassandraDecodeError(obj interface{}, err error) {
	un := obj.(*unstructured.Unstructured)
	log.Errorf("Could not decode Cassandra %s.%s: %v", un.GetNamespace(), un.GetName(), err)
}

func (o *Operator) clusterDeleted(obj interface{}) {
	clusterDefinition, err := unstructuredToCassandra(obj)
	if err != nil {
		logCassandraDecodeError(obj, err)
		return
	}
	v1alpha1helpers.SetDefaultsForCassandra(clusterDefinition)
	o.adjustUseEmptyDir(clusterDefinition)

	clusterID := fmt.Sprintf("%s.%s", clusterDefinition.Namespace, clusterDefinition.Name)
	o.eventDispatcher.Dispatch(&dispatcher.Event{Kind: operations.DeleteCluster, Key: clusterID, Data: clusterDefinition})
}

func (o *Operator) clusterUpdated(old interface{}, new interface{}) {
	oldCluster, err := unstructuredToCassandra(old)
	if err != nil {
		logCassandraDecodeError(old, err)
		return
	}

	newCluster, err := unstructuredToCassandra(new)
	if err != nil {
		logCassandraDecodeError(new, err)
		return
	}

	log.Debug(spew.Sprintf("Cluster update detected for %s.%s, old: %+v \nnew: %+v", oldCluster.Namespace, oldCluster.Name, oldCluster.Spec, newCluster.Spec))

	v1alpha1helpers.SetDefaultsForCassandra(oldCluster)
	o.adjustUseEmptyDir(oldCluster)
	v1alpha1helpers.SetDefaultsForCassandra(newCluster)
	o.adjustUseEmptyDir(newCluster)

	if reflect.DeepEqual(oldCluster.Spec, newCluster.Spec) {
		log.Debugf("update event received for cluster %s.%s but no changes detected", newCluster.Namespace, newCluster.Name)
		return
	}

	clusterID := fmt.Sprintf("%s.%s", newCluster.Namespace, newCluster.Name)
	o.eventDispatcher.Dispatch(&dispatcher.Event{
		Kind: operations.UpdateCluster,
		Key:  clusterID,
		Data: operations.ClusterUpdate{OldCluster: oldCluster, NewCluster: newCluster},
	})
}

func (o *Operator) adjustUseEmptyDir(cluster *v1alpha1.Cassandra) {
	if *cluster.Spec.UseEmptyDir && !o.config.AllowEmptyDir {
		log.Warnf("Cluster %s.%s cannot be configured to use emptyDir, as the operator is configured not to allow the creation of clusters which use emptyDir storage.", cluster.Namespace, cluster.Name)
		cluster.Spec.UseEmptyDir = ptr.Bool(false)
	}
}

func (o *Operator) addSignalHandler(stopCh chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range c {
			log.Infof("Signalled %v, shutting down gracefully", sig)
			close(stopCh)
			os.Exit(0)
		}
	}()
}

func (o *Operator) startServer(metricsPoller *metrics.PrometheusMetrics) {
	o.startMetricPolling(metricsPoller)
	statusCheck := newStatusCheck()
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/live", livenessAndReadinessCheck)
	http.HandleFunc("/ready", livenessAndReadinessCheck)
	http.HandleFunc("/status", statusCheck.statusPage)
	go func() {
		log.Error(http.ListenAndServe(":9090", nil))
		os.Exit(0)
	}()
}

func (o *Operator) startMetricPolling(metricsPoller *metrics.PrometheusMetrics) {
	go func() {
		for {
			for _, c := range o.clusters {
				if c.Online {
					log.Debugf("Sending request for metrics for cluster %s", c.QualifiedName())
					o.eventDispatcher.Dispatch(&dispatcher.Event{Kind: operations.GatherMetrics, Key: c.QualifiedName(), Data: c})
				}
			}
			time.Sleep(o.config.MetricPollInterval)
		}
	}()
}

func livenessAndReadinessCheck(resp http.ResponseWriter, _ *http.Request) {
	resp.WriteHeader(http.StatusNoContent)
}
