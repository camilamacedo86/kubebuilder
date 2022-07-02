/*
Copyright 2022 The Kubernetes authors.

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

package controllers

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	examplecomv1alpha1 "sigs.k8s.io/kubebuilder/testdata/project-v3-with-deploy-image/api/v1alpha1"
)

var _ = Describe("Busybox controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		BusyboxName      = "test-busybox"
		BusyboxNamespace = "default"
		DeploymentName   = "test-deployment"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When running the Busybox controller", func() {
		It("Status of the created deployment should be Running", func() {
			By("By creating a new Busybox")
			ctx := context.Background()
			busybox := &examplecomv1alpha1.Busybox{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "example.com.testproject.org/v1alpha1",
					Kind:       "Busybox",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      BusyboxName,
					Namespace: BusyboxNamespace,
				},
				Spec: examplecomv1alpha1.BusyboxSpec{
					Size: 1,
				},
			}
			Expect(k8sClient.Create(ctx, busybox)).Should(Succeed())

			BusyboxLookupKey := types.NamespacedName{Name: BusyboxName, Namespace: BusyboxNamespace}
			createdBusybox := &examplecomv1alpha1.Busybox{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, BusyboxLookupKey, createdBusybox)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("By checking the Busybox have one active deployment")
			deployment := &appsv1.Deployment{}
			Eventually(
				getResourceFunc(ctx, client.ObjectKey{Name: BusyboxName, Namespace: BusyboxNamespace}, deployment),
				duration, interval).Should(BeNil())
		})
	})

})
