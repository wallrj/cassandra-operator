package webhook

import (
	"context"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1"
	"github.com/sky-uk/cassandra-operator/cassandra-operator/pkg/apis/cassandra/v1alpha1/validation"
)

// +kubebuilder:webhook:path=/validate-v1alpha1-cassandra,mutating=false,failurePolicy=fail,groups="core.sky.uk",resources=cassandras,verbs=create;update,versions=v1alpha1,name=vcass.core.sky.uk

// CassandraValidator validates Cassandras
type CassandraValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// Handle admits a pod iff a specific annotation exists.
func (v *CassandraValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	cass := &v1alpha1.Cassandra{}
	err := v.decoder.Decode(req, cass)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if req.OldObject == nil {
		// return admission.Denied("toot toot")
		return admission.Allowed("")
	}

	oldCass := &v1alpha1.Cassandra{}
	err = v.decoder.DecodeRaw(req.OldObject, oldCass)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	err := validation.ValidateCassandraClusterUpdate(oldCass, cass).ToAggregate()
	if err != nil {
		return admission.Denied(err.Error())
	}
	return admission.Allowed("")
}

// CassandraValidator implements inject.Client.
// A client will be automatically injected.

// InjectClient injects the client.
func (v *CassandraValidator) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// CassandraValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *CassandraValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
