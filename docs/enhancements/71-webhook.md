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
3. Also requires setting up RBAC policy for webhook server to make SubjectAccessReviews back to the API server.


## Ease of writing validation

Controller Tools
1. Easy, just compare old and new structs

Openshift Generic Webhook
1. Easy, just compare old and new structs

## Discussion

A summary of a discussion from #kubebuilder channel in Kubernetes Slack:
 * https://kubernetes.slack.com/archives/CAR30FCJZ/p1561564919014200

richardw [17:01] wrote:
 @deads2k and @directxman12 I'm trying to decide between using https://github.com/openshift/generic-admission-server, https://github.com/kubernetes-sigs/controller-runtime or https://github.com/operator-framework/operator-sdk/issues/1217 to implement a validating admission webhook.
 generic-admission-server seems to be more mature and provides mutual authentication between API server and webhook server.
 Did you consider using that same aggregate API server mechanism for the controller-runtime webhooks?

deads2k [17:13] wrote:
 @richardw I don't plan to try to force people onto "the one true path".  We developed the generic-admission-server a couple years back to make it easy to develop secure webhooks with zero config clients.  It has worked out well overall, but I'm not the sort of person to tell someone not to use another library just because I didn't build it.

richardw [17:18] wrote:
 Understood. The reason I'm asking here, is that I wrote that readme for https://github.com/openshift/generic-admission-server in an effort to persuade a colleague that we should use it.  Still haven't managed to convince them though....the part I still can't explain is why the webhook server needs to send  subjectaccessreviews back to the API server. (edited)
richardw [17:25]
 I didn't make myself clear....what I don't fully understand is the threat that is prevented by doing subjectaccessreviews / RBAC access control on the webhook requests.

Openshift:

deads2k [18:42] wrote:
 @richardw Admission plugins often combine user input with the state of the cluster to make decisions.  Sometimes that state can leak through.  Consider an admission plugin like the namespacelifecycle plugin.  It decides if you can create a resource in a namespace based on whether the namespace exists. Namespace names themselves aren't published (not exposed to all users).  This prevents users from seeing the "people we're going to fire in 2020"  namespace.  If it was an unsecured admission webhook, users would be able to start hunting for namespaces using it.  There are many examples of PII escapes via admission if it were undisclosed.

Controller Tools:

directxman12 [18:57] wrote:
 @richardw re the operator-sdk issue, that'll probably just be controller-runtime at this point, since operatorsdk is based on CR now (edited)
 re: controller-runtime vs generic-admission-server, the interfaces we expose for low-level admission hooks look similar (compare https://github.com/kubernetes-sigs/controller-runtime/blob/master/examples/builtins/validatingwebhook.go)
 we've got a few helpers for higher-level tasks (e.g. writing validating webhooks that are for validating custom resources to augment declarative validation)

 generic-admission-server has certainly been around for longer though, and the approach with aggregated API servers is pretty useful for securing webhooks.  It wouldn't be too hard to set up in CR, I think.  We haven't (as of yet) considered it in CR, mainly because there haven't been very many people asking for it, and we've been focused a lot more on the usecase of "we need to validate that this object is correct to augment declarative validation" and similar
