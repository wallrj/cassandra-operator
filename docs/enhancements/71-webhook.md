# Comparison of ﻿Webhook Libraries

We have evaluated two webhook libraries:

1. Generic Admission Server: https://github.com/openshift/generic-admission-server
2. Controller Runtime: https://github.com/kubernetes-sigs/controller-runtime

Here is our evaluation / comparison:

## Popularity

Openshift
* Used by 10-20 projects
* https://github.com/search?l=Go&q=%22generic-admission-server%22&type=Code
Controller tools
* Used by 5-10  projects
* https://github.com/search?q=%22%2Bkubebuilder%3Awebhook%22&type=Code


## Stability

Controller-tools:
* API still in flux:
   * for example: https://github.com/kubernetes-sigs/controller-runtime/pull/497

Openshift:
* Established October 2017.
* Developed by Redhat and used throughout Openshift.
* Stable.


## Maintainability

Controller tools:
1. Currently unstable.
2. API changes frequently.
3. But likely to be better maintained in future.

Openshift
1. Stable, mostly unchanged for the last 2 years
2. In maintenance mode, mostly only receiving updates for compatibility with new kubernetes versions


## Ease of deployment

Controller tools:
1. Requires certificate management (e.g. Cert-Manager) to be deployed to rotate webhook server certificates.
2. Also requires you to figure out your own API server > Webhook server client authentication.

Openshift
1. Requires certificate management (e.g. Cert-Manager) to be deployed to rotate webhook server certificates.
2. API Server > Webhook server client authentication is handled for you by a well established token rotation system used for aggregate API servers.
3. Also requires setting up RBAC policy for webhook server to make SubjectAccessReviews back to the API server… which may not add any value.


## Ease of writing validation

Controller Tools
1. Easy, just compare old and new structs

Openshift Generic Webhook
1. Easy, just compare old and new structs
