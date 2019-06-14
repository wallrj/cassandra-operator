package main

import (
	"context"
	"os"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
)

var log = logf.Log.WithName("example-controller")

type reconcileCassandra struct {
	client client.Client
	log    logr.Logger
}

var _ reconcile.Reconciler = &reconcileCassandra{}

func (r *reconcileCassandra) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log := r.log.WithValues("request", request)

	// Fetch the Cassandra from the cache
	cass := &v1alpha1.Cassandra{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cass)
	if errors.IsNotFound(err) {
		log.Error(nil, "Could not find Cassandra")
		return reconcile.Result{}, nil
	}

	if err != nil {
		log.Error(err, "Could not fetch Cassandra")
		return reconcile.Result{}, err
	}

	// Print the ReplicaSet
	log.Info("Reconciling Cassandra", "Pod", cass.Spec.Pod)

	// // Set the label if it is missing
	// if rs.Labels == nil {
	//	rs.Labels = map[string]string{}
	// }
	// if rs.Labels["hello"] == "world" {
	//	return reconcile.Result{}, nil
	// }

	// // Update the ReplicaSet
	// rs.Labels["hello"] = "world"
	// err = r.client.Update(context.TODO(), rs)
	// if err != nil {
	//	log.Error(err, "Could not write ReplicaSet")
	//	return reconcile.Result{}, err
	// }

	return reconcile.Result{}, nil
}

func main() {
	logf.SetLogger(zap.Logger(false))
	entryLog := log.WithName("entrypoint")
	entryLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		entryLog.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	// Setup a new controller to reconcile Cassandras
	entryLog.Info("Setting up controller")
	c, err := controller.New("cassandra-controller", mgr, controller.Options{
		Reconciler: &reconcileCassandra{client: mgr.GetClient(), log: log.WithName("reconciler")},
	})
	if err != nil {
		entryLog.Error(err, "unable to set up individual controller")
		os.Exit(1)
	}

	// Watch Cassandras and enqueue ReplicaSet object key
	if err := c.Watch(&source.Kind{Type: &v1alpha1.Cassandra{}}, &handler.EnqueueRequestForObject{}); err != nil {
		entryLog.Error(err, "unable to watch Cassandra")
		os.Exit(1)
	}

	entryLog.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		entryLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
	return
}
