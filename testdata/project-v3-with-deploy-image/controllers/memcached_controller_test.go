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

var _ = Describe("Memcached controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		MemcachedName      = "test-memcached"
		MemcachedNamespace = "default"
		DeploymentName     = "test-deployment"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When running the Memcached controller", func() {
		It("Status of the created deployment should be Running", func() {
			By("By creating a new Memcached")
			ctx := context.Background()
			memcached := &examplecomv1alpha1.Memcached{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "example.com.testproject.org/v1alpha1",
					Kind:       "Memcached",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      MemcachedName,
					Namespace: MemcachedNamespace,
				},
				Spec: examplecomv1alpha1.MemcachedSpec{
					Size: 1,
				},
			}
			Expect(k8sClient.Create(ctx, memcached)).Should(Succeed())

			MemcachedLookupKey := types.NamespacedName{Name: MemcachedName, Namespace: MemcachedNamespace}
			createdMemcached := &examplecomv1alpha1.Memcached{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, MemcachedLookupKey, createdMemcached)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("By checking the Memcached have one active deployment")
			deployment := &appsv1.Deployment{}
			Eventually(
				getResourceFunc(ctx, client.ObjectKey{Name: MemcachedName, Namespace: MemcachedNamespace}, deployment),
				duration, interval).Should(BeNil())
		})
	})

})
