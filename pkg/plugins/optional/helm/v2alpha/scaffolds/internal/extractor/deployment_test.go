/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package extractor

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	keyReplicas   = "replicas"
	keySecretName = "secretName"
	keyConfigMap  = "configMap"
	keyImage      = "image"

	valAppsV1       = "apps/v1"
	valDeployment   = "Deployment"
	valTestDeploy   = "test-deployment"
	valManager      = "manager"
	valAppConfig    = "app-config"
	valAppSecret    = "app-secret"
	valMyConfig     = "my-config"
	valMySecret     = "my-secret"
	valWebhookCert  = "webhook-certs"
	valMetricsCert  = "metrics-certs"
	valMountConfig  = "/etc/config"
	valMountCerts   = "/certs"
	valMountMetrics = "/metrics"
)

type deploymentOpts struct {
	containers  []map[string]any
	volumes     []any
	annotations map[string]string
}

func makeDeployment(opts deploymentOpts) *unstructured.Unstructured {
	podSpec := map[string]any{}
	if opts.containers != nil {
		cs := make([]any, len(opts.containers))
		for i, c := range opts.containers {
			cs[i] = c
		}
		podSpec["containers"] = cs
	}
	if opts.volumes != nil {
		podSpec["volumes"] = opts.volumes
	}
	template := map[string]any{"spec": podSpec}
	if opts.annotations != nil {
		template["metadata"] = map[string]any{
			"annotations": opts.annotations,
		}
	}
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": valAppsV1,
			"kind":       valDeployment,
			"metadata":   map[string]any{"name": valTestDeploy},
			"spec": map[string]any{
				"template": template,
			},
		},
	}
}

func podSpec(d *unstructured.Unstructured) map[string]any {
	return d.Object["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)
}

var _ = Describe("DeploymentExtractor", func() {
	Describe("ExtractDeploymentConfig replicas handling", func() {
		var (
			deployment *unstructured.Unstructured
			extractor  *DeploymentExtractor
		)

		BeforeEach(func() {
			deployment = makeDeployment(deploymentOpts{
				containers: []map[string]any{{
					"name": valManager, keyImage: "controller:latest",
				}},
			})
			extractor = &DeploymentExtractor{}
		})

		It("should return nil when replicas is not set", func() {
			result := extractor.ExtractDeploymentConfig(deployment)
			Expect(result.Manager.Replicas).To(BeNil())
		})

		It("should preserve scale-to-zero", func() {
			deployment.Object["spec"].(map[string]any)[keyReplicas] = int64(0)
			result := extractor.ExtractDeploymentConfig(deployment)
			Expect(result.Manager.Replicas).NotTo(BeNil())
			Expect(*result.Manager.Replicas).To(Equal(0))
		})

		It("should extract replicas value", func() {
			deployment.Object["spec"].(map[string]any)[keyReplicas] = int64(3)
			result := extractor.ExtractDeploymentConfig(deployment)
			Expect(result.Manager.Replicas).NotTo(BeNil())
			Expect(*result.Manager.Replicas).To(Equal(3))
		})
	})

	Describe("ExtractPortFromArg", func() {
		DescribeTable("valid formats",
			func(arg string, expected int) {
				Expect(ExtractPortFromArg(arg)).To(Equal(expected))
			},
			Entry(":PORT", "--metrics-bind-address=:8443", 8443),
			Entry("HOST:PORT", "--metrics-bind-address=0.0.0.0:8080", 8080),
			Entry("localhost:PORT", "--metrics-bind-address=127.0.0.1:9090", 9090),
			Entry("IPv6", "--webhook-port=[::1]:9443", 9443),
			Entry("min valid port", "--metrics-bind-address=:1", 1),
			Entry("max valid port", "--metrics-bind-address=:65535", 65535),
			Entry("multiple colons (IPv6)", "--metrics-bind-address=::1:8443", 8443),
		)

		DescribeTable("invalid formats",
			func(arg string) {
				Expect(ExtractPortFromArg(arg)).To(Equal(0))
			},
			Entry("missing equals", "--metrics-bind-address:8443"),
			Entry("non-numeric", "--metrics-bind-address=:abc"),
			Entry("too low", "--metrics-bind-address=:0"),
			Entry("too high", "--metrics-bind-address=:99999"),
			Entry("empty port", "--metrics-bind-address=:"),
			Entry("missing value", "--metrics-bind-address="),
		)
	})

	Describe("ExtractDeploymentConfig extraVolumes extraction", func() {
		It("should extract custom volumes without mutating the deployment", func() {
			deployment := makeDeployment(deploymentOpts{
				volumes: []any{
					map[string]any{"name": valAppConfig, keyConfigMap: map[string]any{"name": valMyConfig}},
					map[string]any{"name": valAppSecret, "secret": map[string]any{keySecretName: valMySecret}},
				},
				containers: []map[string]any{{
					"name": valManager,
					"volumeMounts": []any{
						map[string]any{"name": valAppConfig, "mountPath": valMountConfig},
						map[string]any{"name": valAppSecret, "mountPath": "/etc/secret"},
					},
				}},
			})

			config := (&DeploymentExtractor{}).ExtractDeploymentConfig(deployment)
			Expect(config.Manager.ExtraVolumes).To(HaveLen(2))
			Expect(config.Manager.ExtraVolumeMounts).To(HaveLen(2))
			Expect(podSpec(deployment)["volumes"].([]any)).To(HaveLen(2), "extraction should not mutate")
		})

		It("should separate system and custom volumes", func() {
			deployment := makeDeployment(deploymentOpts{
				volumes: []any{
					map[string]any{"name": valWebhookCert, "secret": map[string]any{keySecretName: "webhook-server-cert"}},
					map[string]any{"name": valMetricsCert, "secret": map[string]any{keySecretName: "metrics-server-cert"}},
					map[string]any{"name": valAppConfig, keyConfigMap: map[string]any{"name": valMyConfig}},
				},
				containers: []map[string]any{{
					"name": valManager,
					"volumeMounts": []any{
						map[string]any{"name": valWebhookCert, "mountPath": valMountCerts},
						map[string]any{"name": valMetricsCert, "mountPath": valMountMetrics},
						map[string]any{"name": valAppConfig, "mountPath": valMountConfig},
					},
				}},
			})

			config := (&DeploymentExtractor{}).ExtractDeploymentConfig(deployment)
			Expect(config.Manager.ExtraVolumes).To(HaveLen(1))
			Expect(config.Manager.ExtraVolumeMounts).To(HaveLen(1))
		})

		It("should return nil when only system volumes exist", func() {
			deployment := makeDeployment(deploymentOpts{
				volumes: []any{
					map[string]any{"name": valWebhookCert},
					map[string]any{"name": valMetricsCert},
				},
				containers: []map[string]any{{
					"name": valManager,
					"volumeMounts": []any{
						map[string]any{"name": valWebhookCert, "mountPath": valMountCerts},
						map[string]any{"name": valMetricsCert, "mountPath": valMountMetrics},
					},
				}},
			})

			config := (&DeploymentExtractor{}).ExtractDeploymentConfig(deployment)
			Expect(config.Manager.ExtraVolumes).To(BeNil())
			Expect(config.Manager.ExtraVolumeMounts).To(BeNil())
		})
	})

	Describe("StripCustomVolumes", func() {
		It("should remove all custom volumes and leave empty slice", func() {
			deployment := makeDeployment(deploymentOpts{
				volumes: []any{
					map[string]any{"name": valAppConfig, keyConfigMap: map[string]any{"name": valMyConfig}},
				},
				containers: []map[string]any{{
					"name": valManager,
					"volumeMounts": []any{
						map[string]any{"name": valAppConfig, "mountPath": valMountConfig},
					},
				}},
			})

			(&DeploymentExtractor{}).StripCustomVolumes(deployment)

			spec := podSpec(deployment)
			Expect(spec["volumes"].([]any)).To(BeEmpty())
			container := spec["containers"].([]any)[0].(map[string]any)
			Expect(container["volumeMounts"].([]any)).To(BeEmpty())
		})

		It("should keep system volumes and strip custom ones", func() {
			deployment := makeDeployment(deploymentOpts{
				volumes: []any{
					map[string]any{"name": valWebhookCert, "secret": map[string]any{keySecretName: "webhook-server-cert"}},
					map[string]any{"name": valMetricsCert, "secret": map[string]any{keySecretName: "metrics-server-cert"}},
					map[string]any{"name": valAppConfig, keyConfigMap: map[string]any{"name": valMyConfig}},
				},
				containers: []map[string]any{{
					"name": valManager,
					"volumeMounts": []any{
						map[string]any{"name": valWebhookCert, "mountPath": valMountCerts},
						map[string]any{"name": valMetricsCert, "mountPath": valMountMetrics},
						map[string]any{"name": valAppConfig, "mountPath": valMountConfig},
					},
				}},
			})

			(&DeploymentExtractor{}).StripCustomVolumes(deployment)

			spec := podSpec(deployment)
			volumes := spec["volumes"].([]any)
			Expect(volumes).To(HaveLen(2))
			Expect(volumes[0].(map[string]any)["name"]).To(Equal(valWebhookCert))
			Expect(volumes[1].(map[string]any)["name"]).To(Equal(valMetricsCert))
		})

		It("should be a no-op when deployment has no volumes", func() {
			deployment := makeDeployment(deploymentOpts{
				containers: []map[string]any{{"name": valManager}},
			})

			Expect(func() {
				(&DeploymentExtractor{}).StripCustomVolumes(deployment)
			}).NotTo(Panic())

			spec := podSpec(deployment)
			Expect(spec["volumes"]).To(BeNil())
		})

		It("should be idempotent", func() {
			deployment := makeDeployment(deploymentOpts{
				volumes: []any{
					map[string]any{"name": valWebhookCert, "secret": map[string]any{keySecretName: "webhook-server-cert"}},
					map[string]any{"name": valAppConfig, keyConfigMap: map[string]any{"name": valMyConfig}},
				},
				containers: []map[string]any{{
					"name": valManager,
					"volumeMounts": []any{
						map[string]any{"name": valWebhookCert, "mountPath": valMountCerts},
						map[string]any{"name": valAppConfig, "mountPath": valMountConfig},
					},
				}},
			})

			de := &DeploymentExtractor{}
			de.StripCustomVolumes(deployment)
			de.StripCustomVolumes(deployment)

			spec := podSpec(deployment)
			volumes := spec["volumes"].([]any)
			Expect(volumes).To(HaveLen(1))
			Expect(volumes[0].(map[string]any)["name"]).To(Equal(valWebhookCert))
		})

		It("should correctly strip after extract in the full pipeline", func() {
			deployment := makeDeployment(deploymentOpts{
				volumes: []any{
					map[string]any{"name": valWebhookCert, "secret": map[string]any{keySecretName: "webhook-server-cert"}},
					map[string]any{"name": valAppConfig, keyConfigMap: map[string]any{"name": valMyConfig}},
				},
				containers: []map[string]any{{
					"name": valManager,
					"volumeMounts": []any{
						map[string]any{"name": valWebhookCert, "mountPath": valMountCerts},
						map[string]any{"name": valAppConfig, "mountPath": valMountConfig},
					},
				}},
			})

			de := &DeploymentExtractor{}
			config := de.ExtractDeploymentConfig(deployment)

			Expect(config.Manager.ExtraVolumes).To(HaveLen(1), "custom volume in values before strip")
			Expect(config.Manager.ExtraVolumeMounts).To(HaveLen(1), "custom volume mount in values before strip")

			de.StripCustomVolumes(deployment)

			spec := podSpec(deployment)
			volumes := spec["volumes"].([]any)
			Expect(volumes).To(HaveLen(1))
			Expect(volumes[0].(map[string]any)["name"]).To(Equal(valWebhookCert),
				"only system volume remains in deployment after strip")

			container := spec["containers"].([]any)[0].(map[string]any)
			mounts := container["volumeMounts"].([]any)
			Expect(mounts).To(HaveLen(1))
			Expect(mounts[0].(map[string]any)["name"]).To(Equal(valWebhookCert),
				"only system volume mount remains in container after strip")
		})
	})

	Describe("Manager container resolution", func() {
		It("should use the default-container annotation to find the manager", func() {
			deployment := makeDeployment(deploymentOpts{
				annotations: map[string]string{
					"kubectl.kubernetes.io/default-container": "controller-test",
				},
				containers: []map[string]any{
					{"name": "controller-test", keyImage: "controller:latest"},
				},
			})

			config := (&DeploymentExtractor{}).ExtractDeploymentConfig(deployment)
			Expect(config.Manager.Image.Repository).To(Equal("controller"))
			Expect(config.Manager.Image.Tag).To(Equal("latest"))
		})

		It("should find the manager when a sidecar is before it", func() {
			deployment := makeDeployment(deploymentOpts{
				annotations: map[string]string{
					"kubectl.kubernetes.io/default-container": valManager,
				},
				containers: []map[string]any{
					{"name": "sidecar", keyImage: "sidecar:v1"},
					{"name": valManager, keyImage: "controller:latest"},
				},
			})

			config := (&DeploymentExtractor{}).ExtractDeploymentConfig(deployment)
			Expect(config.Manager.Image.Repository).To(Equal("controller"))
		})

		It("should fall back to 'manager' name when annotation is missing", func() {
			deployment := makeDeployment(deploymentOpts{
				containers: []map[string]any{
					{"name": "sidecar", keyImage: "sidecar:v1"},
					{"name": valManager, keyImage: "controller:latest"},
				},
			})

			config := (&DeploymentExtractor{}).ExtractDeploymentConfig(deployment)
			Expect(config.Manager.Image.Repository).To(Equal("controller"))
		})

		It("should fall back to containers[0] when no name matches", func() {
			deployment := makeDeployment(deploymentOpts{
				containers: []map[string]any{
					{"name": "custom-name", keyImage: "controller:latest"},
				},
			})

			config := (&DeploymentExtractor{}).ExtractDeploymentConfig(deployment)
			Expect(config.Manager.Image.Repository).To(Equal("controller"))
		})

		It("should strip volumes from the correct container with sidecar before manager", func() {
			deployment := makeDeployment(deploymentOpts{
				annotations: map[string]string{
					"kubectl.kubernetes.io/default-container": valManager,
				},
				volumes: []any{
					map[string]any{"name": valWebhookCert, "secret": map[string]any{keySecretName: "webhook-server-cert"}},
					map[string]any{"name": valAppConfig, keyConfigMap: map[string]any{"name": valMyConfig}},
				},
				containers: []map[string]any{
					{
						"name": "sidecar",
						"volumeMounts": []any{
							map[string]any{"name": "sidecar-vol", "mountPath": "/sidecar"},
						},
					},
					{
						"name": valManager,
						"volumeMounts": []any{
							map[string]any{"name": valWebhookCert, "mountPath": valMountCerts},
							map[string]any{"name": valAppConfig, "mountPath": valMountConfig},
						},
					},
				},
			})

			(&DeploymentExtractor{}).StripCustomVolumes(deployment)

			spec := podSpec(deployment)
			volumes := spec["volumes"].([]any)
			Expect(volumes).To(HaveLen(1))
			Expect(volumes[0].(map[string]any)["name"]).To(Equal(valWebhookCert))

			sidecar := spec["containers"].([]any)[0].(map[string]any)
			Expect(sidecar["volumeMounts"].([]any)).To(HaveLen(1), "sidecar mounts should be untouched")

			manager := spec["containers"].([]any)[1].(map[string]any)
			managerMounts := manager["volumeMounts"].([]any)
			Expect(managerMounts).To(HaveLen(1))
			Expect(managerMounts[0].(map[string]any)["name"]).To(Equal(valWebhookCert))
		})
	})
})
