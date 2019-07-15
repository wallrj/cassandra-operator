package webhook

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	v1alpha1helpers "github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1/helpers"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1alpha1-cassandra,mutating=true,failurePolicy=fail,groups="core.sky.uk",resources=cassandras,verbs=create;update,versions=v1alpha1,name=defaulter.admission.cassandras.core.sky.uk

// CassandraDefaulter applies defaults to Cassandras
type CassandraDefaulter struct {
	client  client.Client
	decoder *admission.Decoder
}

// Handle applied defaults to the Cassandra
func (v *CassandraDefaulter) Handle(ctx context.Context, req admission.Request) admission.Response {
	cass := &v1alpha1.Cassandra{}

	err := v.decoder.Decode(req, cass)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	v1alpha1helpers.SetDefaultsForCassandra(cass)

	marshaledCass, err := json.Marshal(cass)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledCass)
}

// CassandraValidator implements inject.Client.
// A client will be automatically injected.

// InjectClient injects the client.
func (v *CassandraDefaulter) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// CassandraValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *CassandraDefaulter) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
