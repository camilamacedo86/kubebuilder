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

package appliers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func networkPolicy(name string) *unstructured.Unstructured {
	np := &unstructured.Unstructured{}
	np.SetAPIVersion("networking.k8s.io/v1")
	np.SetKind("NetworkPolicy")
	np.SetName(name)
	return np
}

var _ = Describe("AddConditionalWrappers", func() {
	It("gates the metrics policy on both networkPolicy.enabled and metrics.enabled", func() {
		result := AddConditionalWrappers("kind: NetworkPolicy\n", networkPolicy("test-project-allow-metrics-traffic"))

		Expect(result).To(ContainSubstring(
			"{{- if and .Values.networkPolicy.enabled .Values.metrics.enabled }}"))
		Expect(result).To(ContainSubstring("{{- end }}"))
	})

	It("gates the webhook policy on both networkPolicy.enabled and webhook.enabled", func() {
		result := AddConditionalWrappers("kind: NetworkPolicy\n", networkPolicy("test-project-allow-webhook-traffic"))

		Expect(result).To(ContainSubstring(
			"{{- if and .Values.networkPolicy.enabled .Values.webhook.enabled }}"))
		Expect(result).To(ContainSubstring("{{- end }}"))
	})

	It("gates a custom policy on networkPolicy.enabled only", func() {
		result := AddConditionalWrappers("kind: NetworkPolicy\n", networkPolicy("test-project-allow-dns-traffic"))

		Expect(result).To(ContainSubstring("{{- if .Values.networkPolicy.enabled }}"))
		Expect(result).NotTo(ContainSubstring(".Values.metrics.enabled"))
		Expect(result).NotTo(ContainSubstring(".Values.webhook.enabled"))
	})

	It("does not wrap a NetworkPolicy served by a non-standard apiVersion", func() {
		np := networkPolicy("custom-policy")
		np.SetAPIVersion("acme.io/v1")

		result := AddConditionalWrappers("kind: NetworkPolicy\n", np)

		Expect(result).NotTo(ContainSubstring(".Values.networkPolicy.enabled"))
	})
})
