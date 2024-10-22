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

package manager

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
)

var _ machinery.Template = &ManagerDeployment{}

// ManagerDeployment scaffolds the manager Deployment for the Helm chart
type ManagerDeployment struct {
	machinery.TemplateMixin

	// Force if true will scaffold the env with the images
	DeployImages bool
	// Force if true allow overwrite the scaffolded file
	Force bool
	// HasWebhooks is true when webhooks were found in the config
	HasWebhooks bool
}

// SetTemplateDefaults sets the default template configuration
func (f *ManagerDeployment) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("dist", "chart", "templates", "manager", "manager.yaml")
	}

	f.TemplateBody = managerDeploymentTemplate

	if f.Force {
		f.IfExistsAction = machinery.OverwriteFile
	} else {
		f.IfExistsAction = machinery.SkipFile
	}

	return nil
}

// nolint:lll
const managerDeploymentTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
  namespace: {{ "{{ .Release.Namespace }}" }}
  labels:
    {{ "{{- include \"chart.labels\" . | nindent 4 }}" }}
    control-plane: controller-manager
spec:
  selector:
    matchLabels:
      {{ "{{- include \"chart.selectorLabels\" . | nindent 6 }}" }}
      control-plane: controller-manager
  template:
    metadata:
      labels:
        {{ "{{- include \"chart.labels\" . | nindent 8 }}" }}
        control-plane: controller-manager
    spec:
      serviceAccountName: {{ "{{ .Values.controllerManager.serviceAccountName }}" }}
      containers:
      - name: manager
        image: {{ "{{ .Values.controllerManager.image.repository }}" }}:{{ "{{ .Values.controllerManager.image.tag }}" }}
        args:
          {{ "{{- range .Values.controllerManager.args }}" }}
          - {{ "{{ . }}" }}
          {{ "{{- end }}" }}
        {{ "{{- if .DeployImages }}" }}
        env:
          {{ "{{- range $key, $value := .Values.controllerManager.env }}" }}
          - name: {{ "{{ $key }}" }}
            value: {{ "{{ $value }}" }}
          {{ "{{- end }}" }}
        {{ "{{- end }}" }}
        livenessProbe:
          {{ "{{- toYaml .Values.controllerManager.livenessProbe | nindent 8 }}" }}
        readinessProbe:
          {{ "{{- toYaml .Values.controllerManager.readinessProbe | nindent 8 }}" }}
        resources:
          {{ "{{- toYaml .Values.controllerManager.resources | nindent 8 }}" }}
        securityContext:
          {{ "{{- toYaml .Values.controllerManager.securityContext | nindent 8 }}" }}
        {{- if .HasWebhooks }}
        volumeMounts:
        - name: webhook-cert
          mountPath: /tmp/k8s-webhook-server/serving-certs
          readOnly: true
        {{- end }}
      terminationGracePeriodSeconds: {{ "{{ .Values.controllerManager.terminationGracePeriodSeconds }}" }}
      {{- if .HasWebhooks }}
      volumes:
      - name: webhook-cert
        secret:
          secretName: webhook-server-cert
      {{- end }}
`
