/*
Copyright 2024 The Kubernetes Authors.

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

package templates

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v4/pkg/plugins/optional/helm"
)

var _ machinery.Template = &HelmValues{}

// HelmValues scaffolds a file that defines the values.yaml structure for the Helm chart
type HelmValues struct {
	machinery.TemplateMixin
	machinery.ProjectNameMixin

	// DeployImages stores the images used for the DeployImage plugin
	DeployImages map[string]string
	// Force if true allows overwriting the scaffolded file
	Force bool
	// Webhooks stores the webhook configurations
	Webhooks []helm.WebhookYAML
	// HasWebhooks is true when webhooks were found in the config
	HasWebhooks bool
}

// SetTemplateDefaults implements file.Template
func (f *HelmValues) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("dist", "chart", "values.yaml")
	}
	f.TemplateBody = helmValuesTemplate

	if f.Force {
		f.IfExistsAction = machinery.OverwriteFile
	} else {
		f.IfExistsAction = machinery.SkipFile
	}

	return nil
}

const helmValuesTemplate = `# [MANAGER]: Manager Deployment Configurations
controllerManager:
  container:
    image:
      repository: controller
      tag: latest
    replicas: 1
    args:
      - "--leader-elect"
      - "--metrics-bind-address=:8443"
      - "--health-probe-bind-address=:8081"
    resources:
      limits:
        cpu: 500m
        memory: 128Mi
      requests:
        cpu: 10m
        memory: 64Mi
    livenessProbe:
      initialDelaySeconds: 15
      periodSeconds: 20
      httpGet:
        path: /healthz
        port: 8081
    readinessProbe:
      initialDelaySeconds: 5
      periodSeconds: 10
      httpGet:
        path: /readyz
        port: 8081
    {{- if .DeployImages }}
    env:
    {{- range $kind, $image := .DeployImages }}
      {{ $kind }}_IMAGE: {{ $image }}
    {{- end }}
    {{- end }}
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
          - "ALL"
  securityContext:
    runAsNonRoot: true
    seccompProfile:
      type: RuntimeDefault
  terminationGracePeriodSeconds: 10
  serviceAccountName: {{ .ProjectName }}-controller-manager

# [RBAC]: To enable RBAC (Permissions) configurations
rbac:
  enable: true

# [CRDs]: To enable the CRDs
crd:
  # This option determines whether the CRDs are included
  # in the installation process.
  enable: true

  # Enabling this option adds the "helm.sh/resource-policy": keep
  # annotation to the CRD, ensuring it remains installed even when
  # the Helm release is uninstalled.
  # NOTE: Removing the CRDs will also remove all cert-manager CR(s)
  # (Certificates, Issuers, ...) due to garbage collection.
  keep: true

# [SAMPLES]: To apply the sample(s) manifests set true
samples:
  enable: false

# [METRICS]: Set to true to generate manifests for exporting metrics.
# To disable metrics export set false, and ensure that the
# ControllerManager argument "--metrics-bind-address=:8443" is removed.
metrics:
  enable: true
{{ if .Webhooks }}
# [WEBHOOKS]: Webhooks configuration
# The following configuration is automatically generated from the manifests
# generated by controller-gen. To update run 'make manifests' and 
# the edit command with the '--force' flag
webhook:
  enable: true
  services:
  {{- range .Webhooks }}
    - name: {{ .Name }}
      type: {{ .Type }}
      path: {{ .Path }}
      failurePolicy: {{ .FailurePolicy }}
      sideEffects: {{ .SideEffects }}
      admissionReviewVersions:
      {{- range .AdmissionReviewVersions }}
        - {{ . }}
      {{- end }}
      rules:
      {{- range .Rules }}
        - operations:
          {{- range .Operations }}
            - {{ . }}
          {{- end }}
          apiGroups:
          {{- if .APIGroups }}
            {{- range .APIGroups }}
              {{- if eq . "" }}
            - ""
              {{- else }}
            - {{ . }}
              {{- end }}
            {{- end }}
          {{- else }}
            - ""
          {{- end }}
          apiVersions:
          {{- range .APIVersions }}
            - {{ . }}
          {{- end }}
          resources:
          {{- range .Resources }}
            - {{ . }}
          {{- end }}
      {{- end }}
  {{- end }}
{{ end }}
# [PROMETHEUS]: To enable a ServiceMonitor to export metrics to Prometheus set true
prometheus:
  enable: false

# [CERT-MANAGER]: To enable cert-manager injection to webhooks set true
certmanager:
  enable: {{ .HasWebhooks }}

# [NETWORK POLICIES]: To enable NetworkPolicies set true
networkPolicy:
  enable: false
`
