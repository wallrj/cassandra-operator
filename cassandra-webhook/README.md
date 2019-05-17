# Cassandra Operator Webhook

A [Kubernetes Admission Webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) which validates `Cassandra` CRD resources and adds default values for optional fields.

This webhook is based on [generic-admission-server by Openshift](https://github.com/openshift/generic-admission-server), a library for writing admission webhooks, which is its self based on the [Kubernetes aggregated API server library](https://github.com/kubernetes/apiserver).

The Kubernetes API communicates with webhook servers using TLS encrypted HTTP connections.
Therefore you must create a TLS private key and signed certificate for the webhook server,
and configure the Kubernetes API server so that it can verify the signature of the webhook server certificate.

Additionally you must configure the webhook server to verify the client certificate of the Kubernetes API server.
This is done

You can deploy the web hook for
