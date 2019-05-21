package operator

import (
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/operator/operations"
	"time"

	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	v1alpha1helpers "github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1/helpers"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/client/clientset/versioned"
	informers "github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/client/informers/externalversions"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/cluster"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/dispatcher"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/metrics"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/util/ptr"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"reflect"
)

// The Operator itself.
type Operator struct {
	clusters           map[string]*cluster.Cluster
	kubeClientset      *kubernetes.Clientset
	cassandraClientset *versioned.Clientset
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
func New(kubeClientset *kubernetes.Clientset, cassandraClientset *versioned.Clientset, operatorConfig *Config) *Operator {
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
	cassandraInformer.Start(o.stopCh)
	configMapInformer.Run(o.stopCh)
	<-o.stopCh
	log.Info("Operator shutting down")
}

func registerCassandraInformer(o *Operator, ns string) informers.SharedInformerFactory {
	cassandraInformerFactory := informers.NewSharedInformerFactoryWithOptions(o.cassandraClientset, resourceResyncInterval, informers.WithNamespace(ns))
	cassandraInformer := cassandraInformerFactory.Core().V1alpha1().Cassandras()
	cassandraInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.clusterAdded,
		UpdateFunc: o.clusterUpdated,
		DeleteFunc: o.clusterDeleted,
	})
	return cassandraInformerFactory
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
	cm := obj.(*v1.ConfigMap)

	if cluster.LooksLikeACassandraConfigMap(cm) {
		clusterID := fmt.Sprintf("%s.%s", cm.Namespace, cm.Name)
		o.eventDispatcher.Dispatch(&dispatcher.Event{Kind: operations.AddCustomConfig, Key: clusterID, Data: cm})
	}
}

func (o *Operator) configMapDeleted(obj interface{}) {
	cm := obj.(*v1.ConfigMap)

	if cluster.ConfigMapBelongsToAManagedCluster(o.clusters, cm) {
		clusterID := fmt.Sprintf("%s.%s", cm.Namespace, cm.Name)
		o.eventDispatcher.Dispatch(&dispatcher.Event{Kind: operations.DeleteCustomConfig, Key: clusterID, Data: cm})
	}
}

func (o *Operator) configMapUpdated(old interface{}, new interface{}) {
	oldConfigMap := old.(*v1.ConfigMap)
	newConfigMap := new.(*v1.ConfigMap)

	if reflect.DeepEqual(oldConfigMap.Data, newConfigMap.Data) {
		log.Debugf("update event received for config map %s.%s but no changes detected", newConfigMap.Namespace, newConfigMap.Name)
		return
	}

	if cluster.ConfigMapBelongsToAManagedCluster(o.clusters, oldConfigMap) {
		clusterID := fmt.Sprintf("%s.%s", oldConfigMap.Namespace, oldConfigMap.Name)
		o.eventDispatcher.Dispatch(&dispatcher.Event{Kind: operations.UpdateCustomConfig, Key: clusterID, Data: newConfigMap})
	}
}

func (o *Operator) clusterAdded(obj interface{}) {
	clusterDefinition := obj.(*v1alpha1.Cassandra)
	o.adjustUseEmptyDir(clusterDefinition)

	clusterID := fmt.Sprintf("%s.%s", clusterDefinition.Namespace, clusterDefinition.Name)
	o.eventDispatcher.Dispatch(&dispatcher.Event{Kind: operations.AddCluster, Key: clusterID, Data: clusterDefinition})
}

func (o *Operator) clusterDeleted(obj interface{}) {
	clusterDefinition := obj.(*v1alpha1.Cassandra)
	o.adjustUseEmptyDir(clusterDefinition)

	clusterID := fmt.Sprintf("%s.%s", clusterDefinition.Namespace, clusterDefinition.Name)
	o.eventDispatcher.Dispatch(&dispatcher.Event{Kind: operations.DeleteCluster, Key: clusterID, Data: clusterDefinition})
}

func (o *Operator) clusterUpdated(old interface{}, new interface{}) {
	oldCluster := old.(*v1alpha1.Cassandra)
	newCluster := new.(*v1alpha1.Cassandra)
	log.Debugf("Cluster update detected for %s.%s, old: %v \nnew: %v", oldCluster.Namespace, oldCluster.Name, oldCluster.Spec, newCluster.Spec)

	o.adjustUseEmptyDir(oldCluster)
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
	if v1alpha1helpers.UseEmptyDir(cluster) && !o.config.AllowEmptyDir {
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
