package main

import (
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	cassandrawebhook "github.com/sky-uk/cassandra-operator/cassandra-webhook/pkg/webhook"
)

func main() {
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.WithError(err).Fatal("failed to set up controller-runtime manager")
	}

	// Setup webhooks
	log.Info("setting up webhook server")
	hookServer := mgr.GetWebhookServer()
	hookServer.Port = 8443
	log.Info("registering webhooks to the webhook server")
	// hookServer.Register("/mutate-v1alpha1-cassandra", &webhook.Admission{Handler: &podAnnotator{}})
	hookServer.Register("/validate-v1alpha1-cassandra", &webhook.Admission{Handler: &cassandrawebhook.CassandraValidator{}})

	log.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.WithError(err).Fatal("unable to run manager")
	}
	return
}
