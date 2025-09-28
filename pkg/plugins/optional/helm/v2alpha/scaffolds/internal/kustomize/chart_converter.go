/*
Copyright 2025 The Kubernetes Authors.

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

package kustomize

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
)

// ChartConverter orchestrates the conversion of kustomize output to Helm chart templates
type ChartConverter struct {
	resources   *ParsedResources
	projectName string
	outputDir   string

	// Components for conversion
	organizer *ResourceOrganizer
	templater *HelmTemplater
	writer    *ChartWriter
	ports     PortsConfig
}

// PortsConfig captures port configuration detected from kustomize manifests.
type PortsConfig struct {
	MetricsPort          string
	WebhookServicePort   string
	WebhookContainerPort string
}

func defaultPortsConfig() PortsConfig {
	return PortsConfig{
		MetricsPort:          "8443",
		WebhookServicePort:   "443",
		WebhookContainerPort: "9443",
	}
}

// NewChartConverter creates a new chart converter with all necessary components
func NewChartConverter(resources *ParsedResources, projectName, outputDir string) *ChartConverter {
	organizer := NewResourceOrganizer(resources)
	templater := NewHelmTemplater(projectName)
	ports := calculatePortsConfig(resources)
	templater.ConfigurePorts(ports)
	writer := NewChartWriter(templater, outputDir)

	return &ChartConverter{
		resources:   resources,
		projectName: projectName,
		outputDir:   outputDir,
		organizer:   organizer,
		templater:   templater,
		writer:      writer,
		ports:       ports,
	}
}

// PortsConfig returns the detected port configuration.
func (c *ChartConverter) PortsConfig() PortsConfig {
	return c.ports
}

// WriteChartFiles converts all resources to Helm chart templates and writes them to the filesystem
func (c *ChartConverter) WriteChartFiles(fs machinery.Filesystem) error {
	// Organize resources by their logical function
	resourceGroups := c.organizer.OrganizeByFunction()

	// Write each group to appropriate template files
	for groupName, resources := range resourceGroups {
		if len(resources) > 0 {
			// De-duplicate exact resources by (apiVersion, kind, namespace, name)
			deduped := dedupeResources(resources)
			if err := c.writer.WriteResourceGroup(fs, groupName, deduped); err != nil {
				return fmt.Errorf("failed to write %s resources: %w", groupName, err)
			}
		}
	}

	return nil
}

// dedupeResources removes exact duplicate resources by keying on
// apiVersion, kind, namespace (optional), and name.
func dedupeResources(resources []*unstructured.Unstructured) []*unstructured.Unstructured {
	seen := make(map[string]struct{})
	out := make([]*unstructured.Unstructured, 0, len(resources))
	for _, r := range resources {
		if r == nil {
			continue
		}
		key := r.GetAPIVersion() + "|" + r.GetKind() + "|" + r.GetNamespace() + "|" + r.GetName()
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, r)
	}
	return out
}

// ExtractDeploymentConfig extracts configuration values from the deployment for values.yaml
func (c *ChartConverter) ExtractDeploymentConfig() map[string]interface{} {
	if c.resources.Deployment == nil {
		return make(map[string]interface{})
	}

	config := make(map[string]interface{})

	// Extract from deployment spec
	spec, found, err := unstructured.NestedFieldNoCopy(c.resources.Deployment.Object, "spec", "template", "spec")
	if !found || err != nil {
		return config
	}

	specMap, ok := spec.(map[string]interface{})
	if !ok {
		return config
	}

	// Extract pod security context
	if podSecurityContext, podSecFound, podSecErr := unstructured.NestedFieldNoCopy(specMap,
		"securityContext"); podSecFound && podSecErr == nil {
		if podSecMap, podSecOk := podSecurityContext.(map[string]interface{}); podSecOk && len(podSecMap) > 0 {
			config["podSecurityContext"] = podSecurityContext
		}
	}

	// Extract container configuration
	containers, found, err := unstructured.NestedFieldNoCopy(specMap, "containers")
	if !found || err != nil {
		return config
	}

	containersList, ok := containers.([]interface{})
	if !ok || len(containersList) == 0 {
		return config
	}

	// Use the first container (manager container)
	firstContainer, ok := containersList[0].(map[string]interface{})
	if !ok {
		return config
	}

	// Extract environment variables
	if env, envFound, envErr := unstructured.NestedFieldNoCopy(firstContainer, "env"); envFound && envErr == nil {
		if envList, envOk := env.([]interface{}); envOk && len(envList) > 0 {
			config["env"] = envList
		}
	}

	// Extract resources
	if resources, resFound, resErr := unstructured.NestedFieldNoCopy(firstContainer,
		"resources"); resFound && resErr == nil {
		if resourcesMap, resOk := resources.(map[string]interface{}); resOk && len(resourcesMap) > 0 {
			config["resources"] = resources
		}
	}

	// Extract container security context
	if securityContext, secFound, secErr := unstructured.NestedFieldNoCopy(firstContainer,
		"securityContext"); secFound && secErr == nil {
		if secMap, secOk := securityContext.(map[string]interface{}); secOk && len(secMap) > 0 {
			config["securityContext"] = securityContext
		}
	}

	return config
}

func calculatePortsConfig(resources *ParsedResources) PortsConfig {
	ports := defaultPortsConfig()

	if resources == nil {
		return ports
	}

	// Extract metrics port from deployment args if available
	if resources.Deployment != nil {
		if containersVal, found, err := unstructured.NestedFieldNoCopy(resources.Deployment.Object,
			"spec", "template", "spec", "containers"); found && err == nil {
			if containers, ok := containersVal.([]interface{}); ok && len(containers) > 0 {
				if container, ok := containers[0].(map[string]interface{}); ok {
					if args, ok := container["args"].([]interface{}); ok {
						for _, arg := range args {
							if s, ok := arg.(string); ok && strings.Contains(s, "--metrics-bind-address=") {
								if idx := strings.LastIndex(s, ":"); idx != -1 && idx+1 < len(s) {
									port := s[idx+1:]
									if port != "" {
										ports.MetricsPort = port
									}
								}
							}
						}
					}
					if cPorts, ok := container["ports"].([]interface{}); ok {
						for _, p := range cPorts {
							pm, ok := p.(map[string]interface{})
							if !ok {
								continue
							}
							if name, ok := pm["name"].(string); ok && strings.Contains(name, "webhook") {
								if val, ok := pm["containerPort"].(int64); ok {
									ports.WebhookContainerPort = fmt.Sprintf("%d", val)
								}
							}
							if val, ok := pm["containerPort"].(int64); ok && len(cPorts) == 1 {
								ports.WebhookContainerPort = fmt.Sprintf("%d", val)
							}
						}
					}
				}
			}
		}
	}

	// Extract metrics and webhook service ports from services
	for _, svc := range resources.Services {
		name := svc.GetName()
		portsVal, found, err := unstructured.NestedFieldNoCopy(svc.Object, "spec", "ports")
		if !found || err != nil {
			continue
		}
		portsList, ok := portsVal.([]interface{})
		if !ok || len(portsList) == 0 {
			continue
		}

		firstPort, ok := portsList[0].(map[string]interface{})
		if !ok {
			continue
		}

		if strings.Contains(name, "metrics") {
			if val, ok := firstPort["port"].(int64); ok {
				ports.MetricsPort = fmt.Sprintf("%d", val)
			}
		}

		if strings.Contains(name, "webhook") {
			if val, ok := firstPort["port"].(int64); ok {
				ports.WebhookServicePort = fmt.Sprintf("%d", val)
			}
			if target, ok := firstPort["targetPort"].(int64); ok {
				ports.WebhookContainerPort = fmt.Sprintf("%d", target)
			}
		}
	}

	return ports
}
