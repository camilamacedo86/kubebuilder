# Available plugins

This section describes the plugins supported and shipped in with the Kubebuilder project.

| Plugin                                                                             | Key                  | Description                                                                                                                                                                                                                                  |
| ---------------------------------------------------------------------------------- | -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [go.kubebuilder.io/v2 - (Deprecated)](go-v2-plugin.md)                             | `go/v2`              | Golang plugin responsible for scaffolding the legacy layout provided with Kubebuilder CLI >= `2.0.0` and < `3.0.0`.                                                                                                                          |
| [go.kubebuilder.io/v3 - (Default scaffold with Kubebuilder init)](go-v3-plugin.md) | `go/v3`              | Default scaffold used for creating a project when no plugin(s) are provided. Responsible for scaffolding Golang projects and its configurations.                                                                                             |
| [declarative.go.kubebuilder.io/v1](declarative-v1.md)                              | `declarative/v1`     | Optional plugin used to scaffold APIs/controllers using the [kubebuilder-declarative-pattern][kubebuilder-declarative-pattern] project.                                                                                                      |
| [kustomize.common.kubebuilder.io/v1](kustomize-v1.md)                              | `kustomize/v1`       | Responsible for scaffold all manifests to configure the projects with [kustomize(v3)][kustomize]. (create and update the the `config/` directory). This plugin is used in the composition to create the plugin (`go/v3`).                    |
| `base.go.kubebuilder.io/v3`                                                        | `base/v3`            | Responsible for scaffold all files which specific requires Golang. This plugin is used in the composition to create the plugin (`go/v3`)                                                                                                     |
| [grafana.kubebuilder.io/v1-alpha](grafana-v1-alpha.md)                             | `grafana/v1-alpha`   | Optional helper plugin which can be used to scaffold Grafana Manifests Dashboards for the default metrics which are exported by controller-runtime.                                                                                                 |
| [deploy-image.go.kubebuilder.io/v1-alpha](deploy-image-plugin-v1-alpha)            | `deploy-image/v1-alpha`   | Optional helper plugin which can be used to scaffold APIs and controller with code implementation to Deploy and Manage an Operand(image).                                                                                                 |

> Note: **ALPHA** plugins can introduce breaking changes. For further info see [Plugins Versioning](./plugins/plugins-versioning.md).

<aside class="note">

<h1>You can also create your own plugins, see:</h1>

- [Creating your own plugins][create-plugins].

</aside>

[create-plugins]: creating-plugins.md
[kubebuilder-declarative-pattern]: https://github.com/kubernetes-sigs/kubebuilder-declarative-pattern
[kustomize]: https://kustomize.io/
