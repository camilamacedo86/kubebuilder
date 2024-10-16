# Webhook

Webhooks are requests for information sent in a blocking fashion. A web
application implementing webhooks will send a HTTP request to other applications
when a certain event happens.

In the kubernetes world, there are 3 kinds of webhooks:
[admission webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks),
[authorization webhook](https://kubernetes.io/docs/reference/access-authn-authz/webhook/) and
[CRD conversion webhook](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/#webhook-conversion).

In [controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/webhook?tab=doc)
libraries, we support admission webhooks and CRD conversion webhooks.

Kubernetes supports these dynamic admission webhooks as of version 1.9 (when the
feature entered beta).

Kubernetes supports the conversion webhooks as of version 1.15 (when the
feature entered beta).

<aside class="note">
<H1>Webhooks Changes from Kubebuilder Release 4.3.0</H1>

Note that before release `4.3.0`, webhooks were scaffolded under the directory `/api` and used the methods
`webhook.Validator` and `webhook.Defaulter`. However, starting from controller-runtime release `v0.20.0`,
these methods are no longer provided, and users should implement custom interfaces.

Examples can be found under the testdata samples: [testdata/project-v4/internal/webhook/v1](https://github.com/kubernetes-sigs/kubebuilder/tree/master/testdata/project-v4/internal/webhook/v1)

Additionally, webhooks should no longer be placed under the `/api` directory.
They should be moved to `internal/webhook`.

You can temporarily scaffold using `--legacy=true`, but this flag will
be removed in future releases.

</aside>