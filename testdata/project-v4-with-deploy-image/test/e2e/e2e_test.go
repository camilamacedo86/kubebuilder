/*
Copyright 2024 The Kubernetes authors.

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

package e2e

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/kubebuilder/testdata/project-v4-with-deploy-image/test/utils"
)

// namespace where the project is deployed in
const namespace = "project-v4-with-deploy-image-system"

// serviceAccountName created for the project
const serviceAccountName = "controller-manager"

// metricsServiceName is the name of the metrics service of the project
const metricsServiceName = "project-v4-with-deploy-image-controller-manager-metrics-service"

// metricsRoleBindingName is the name of the RBAC that will be created to allow get the metrics data
const metricsRoleBindingName = "project-v4-with-deploy-image-metrics-binding"

var _ = Describe("Manager", Ordered, func() {
	// Before running the tests, set up the environment by creating the namespace,
	// installing CRDs, and deploying the controller.
	BeforeAll(func() {
		By("creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", namespace)
		_, err := utils.Run(cmd)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Failed to create namespace")

		By("installing CRDs")
		cmd = exec.Command("make", "install")
		_, err = utils.Run(cmd)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Failed to install CRDs")

		By("deploying the controller-manager")
		cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectImage))
		_, err = utils.Run(cmd)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Failed to deploy the controller-manager")
	})

	// After all tests have been executed, clean up by undeploying the controller, uninstalling CRDs,
	// and deleting the namespace.
	AfterAll(func() {
		By("cleaning up the curl pod for metrics")
		cmd := exec.Command("kubectl", "delete", "pod", "curl-metrics", "-n", namespace)
		_, _ = utils.Run(cmd)

		By("undeploying the controller-manager")
		cmd = exec.Command("make", "undeploy")
		_, _ = utils.Run(cmd)

		By("uninstalling CRDs")
		cmd = exec.Command("make", "uninstall")
		_, _ = utils.Run(cmd)

		By("removing manager namespace")
		cmd = exec.Command("kubectl", "delete", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	// The Context block contains the actual tests that validate the manager's behavior.
	Context("Manager", func() {
		It("should run successfully", func() {
			var controllerPodName string

			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func() error {
				// Get the name of the controller-manager pod
				cmd := exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)

				podOutput, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred(), "Failed to retrieve controller-manager pod information")
				podNames := utils.GetNonEmptyLines(string(podOutput))
				if len(podNames) != 1 {
					return fmt.Errorf("expected 1 controller pod running, but got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))

				// Validate the pod's status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				status, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred(), "Failed to retrieve controller-manager pod status")
				if string(status) != "Running" {
					return fmt.Errorf("controller pod in %s status", status)
				}
				return nil
			}
			// Repeatedly check if the controller-manager pod is running until it succeeds or times out.
			EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())
		})

		It("should ensure the metrics endpoint is serving metrics", func() {
			// Step 1: Create a ClusterRoleBinding for the service account to allow access to metrics
			cmd := exec.Command("kubectl", "create", "clusterrolebinding", metricsRoleBindingName,
				"--clusterrole=metrics-reader",
				"--serviceaccount="+namespace+":"+serviceAccountName)
			_, err := utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Failed to create ClusterRoleBinding")

			// Step 2: Ensure the metrics service is available
			By("validating that the metrics service is available")
			cmd = exec.Command("kubectl", "get", "service", metricsServiceName, "-n", namespace)
			_, err = utils.Run(cmd)
			ExpectWithOffset(2, err).NotTo(HaveOccurred(), "Metrics service should exist")

			// Step 3: Get the service account token
			token, err := serviceAccountToken()
			ExpectWithOffset(2, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, token).NotTo(BeEmpty())

			// Step 4: Create a curl pod to access the metrics endpoint using the token
			By("creating a curl pod to access the metrics endpoint")
			cmd = exec.Command("kubectl", "run", "curl-metrics", "--restart=Never",
				"--namespace", namespace,
				"--image=curlimages/curl:7.78.0",
				"--", "/bin/sh", "-c", fmt.Sprintf(
					"curl -v -k -H 'Authorization: Bearer %s' https://%s.%s.svc.cluster.local:8443/metrics",
					token, metricsServiceName, namespace))
			_, err = utils.Run(cmd)
			ExpectWithOffset(2, err).NotTo(HaveOccurred(), "Failed to create curl pod")

			// Step 5: Wait for the curl pod to complete. We need to use the pod to get the metrics data
			By("validating that the curl pod is running as expected")
			verifyCurlUp := func() error {
				cmd := exec.Command("kubectl", "get", "pods", "curl-metrics",
					"-o", "jsonpath={.status.phase}",
					"-n", namespace)
				status, err := utils.Run(cmd)
				ExpectWithOffset(3, err).NotTo(HaveOccurred())
				if string(status) != "Succeeded" {
					return fmt.Errorf("curl pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(2, verifyCurlUp, 240*time.Second, time.Second).Should(Succeed())

			// Step 6: Retrieve and validate the metrics output
			By("getting the curl-metrics logs")
			metricsOutput := getMetricsOutput()
			ExpectWithOffset(1, metricsOutput).To(ContainSubstring(
				"controller_runtime_reconcile_total",
			))
		})

		// TODO: Customize the e2e test suite with scenarios specific to your project.
		// Consider applying sample/CR(s) and check their status and/or verifying
		// the reconciliation by using the metrics, i.e.:
		// metricsOutput := getMetricsOutput()
		// ExpectWithOffset(1, metricsOutput).To(ContainSubstring(
		//    fmt.Sprintf(`controller_runtime_reconcile_total{controller="%s",result="success"} 1`,
		//    strings.ToLower(<Kind>),
		// )))
	})
})

// serviceAccountToken returns a token for the specified service account in the given namespace.
// It uses the Kubernetes TokenRequest API to generate a token by directly sending a request
// and parsing the resulting token from the API response.
func serviceAccountToken() (string, error) {
	const tokenRequestRawString = `{
		"apiVersion": "authentication.k8s.io/v1",
		"kind": "TokenRequest"
	}`

	// Execute kubectl command to create the token
	cmd := exec.Command("kubectl", "create", "--raw", fmt.Sprintf(
		"/api/v1/namespaces/%s/serviceaccounts/%s/token",
		namespace,
		serviceAccountName,
	), "-f", "-")
	cmd.Stdin = strings.NewReader(tokenRequestRawString)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute kubectl command: %v, output: %s", err, string(output))
	}

	// Parse the JSON output to extract the token
	var token tokenRequest
	if err := json.Unmarshal(output, &token); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON response: %v", err)
	}

	return token.Status.Token, nil
}

// getMetricsOutput retrieves and returns the logs from the curl pod used to access the metrics endpoint.
func getMetricsOutput() string {
	By("getting the curl-metrics logs")
	cmd := exec.Command("kubectl", "logs", "curl-metrics", "-n", namespace)
	metricsOutput, err := utils.Run(cmd)
	ExpectWithOffset(3, err).NotTo(HaveOccurred(), "Failed to retrieve logs from curl pod")
	metricsOutputStr := string(metricsOutput)
	ExpectWithOffset(3, metricsOutputStr).To(ContainSubstring("< HTTP/1.1 200 OK"))
	return metricsOutputStr
}

// tokenRequest is a simplified representation of the Kubernetes TokenRequest API response,
// containing only the token field that we need to extract.
type tokenRequest struct {
	Status struct {
		Token string `json:"token"`
	} `json:"status"`
}
