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
)

var _ = Describe("TemplatePorts NetworkPolicy", func() {
	const metricsPolicy = `spec:
  policyTypes:
    - Ingress
  ingress:
    - from:
      - namespaceSelector:
          matchLabels:
            metrics: enabled
      ports:
        - port: 8443
          protocol: TCP
`

	It("templates the metrics policy ingress port", func() {
		result := TemplatePorts(metricsPolicy, networkPolicy("test-project-allow-metrics-traffic"))

		Expect(result).To(ContainSubstring("port: {{ .Values.metrics.port }}"))
		Expect(result).NotTo(ContainSubstring("port: 8443"))
	})

	It("templates the webhook policy ingress port", func() {
		webhookPolicy := `spec:
  policyTypes:
    - Ingress
  ingress:
    - from:
      - namespaceSelector:
          matchLabels:
            webhook: enabled
      ports:
        - port: 9443
          protocol: TCP
`
		result := TemplatePorts(webhookPolicy, networkPolicy("test-project-allow-webhook-traffic"))

		Expect(result).To(ContainSubstring("port: {{ .Values.webhook.port }}"))
		Expect(result).NotTo(ContainSubstring("port: 9443"))
	})

	It("leaves a custom policy's ports untouched", func() {
		dnsPolicy := `spec:
  policyTypes:
    - Ingress
  ingress:
    - ports:
        - port: 5353
          protocol: UDP
`
		result := TemplatePorts(dnsPolicy, networkPolicy("test-project-allow-dns-traffic"))

		Expect(result).To(ContainSubstring("port: 5353"))
		Expect(result).NotTo(ContainSubstring(".Values."))
	})

	It("templates only the ingress port and preserves egress ports of a metrics policy", func() {
		mixedPolicy := `spec:
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - ports:
        - port: 8443
          protocol: TCP
  egress:
    - ports:
        - port: 53
          protocol: UDP
`
		result := TemplatePorts(mixedPolicy, networkPolicy("test-project-allow-metrics-traffic"))

		Expect(result).To(ContainSubstring("port: {{ .Values.metrics.port }}"))
		Expect(result).NotTo(ContainSubstring("port: 8443"))
		Expect(result).To(ContainSubstring("port: 53"), "egress port must not be rewritten")
	})
})
