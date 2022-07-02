/*
Copyright 2022 The Kubernetes Authors.

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
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ControllerTest{}

// ControllerTest scaffolds the file that defines tests for the controller for a CRD or a builtin resource
// nolint:maligned
type ControllerTest struct {
	machinery.TemplateMixin
	machinery.MultiGroupMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin

	Image string
}

// SetTemplateDefaults implements file.Template
func (f *ControllerTest) SetTemplateDefaults() error {
	if f.Path == "" {
		if f.MultiGroup && f.Resource.Group != "" {
			f.Path = filepath.Join("controllers", "%[group]", "%[kind]_controller_test.go")
		} else {
			f.Path = filepath.Join("controllers", "%[kind]_controller_test.go")
		}
	}
	f.Path = f.Resource.Replacer().Replace(f.Path)

	fmt.Println("creating import for %", f.Resource.Path)
	f.TemplateBody = controllerTestTemplate

	// This one is to overwrite the controller if it exist
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

//nolint:lll
const controllerTestTemplate = `{{ .Boilerplate }}

package {{ if and .MultiGroup .Resource.Group }}{{ .Resource.PackageName }}{{ else }}controllers{{ end }}

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	
	{{ if not (isEmptyStr .Resource.Path) -}}
	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	{{- end }}
)

var _ = Describe("{{ .Resource.Kind }} controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		{{ .Resource.Kind }}Name      = "test-{{ lower .Resource.Kind }}"
		{{ .Resource.Kind }}Namespace = "default"
		DeploymentName = "test-deployment"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When running the {{ .Resource.Kind }} controller", func() {
		It("Status of the created deployment should be Running", func() {
			By("By creating a new {{ .Resource.Kind }}")
			ctx := context.Background()
			{{ lower .Resource.Kind }} := &{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "{{ .Resource.Group }}.{{ .Resource.Domain }}/{{ .Resource.Version }}",
					Kind:       "{{ .Resource.Kind }}",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      {{ .Resource.Kind }}Name,
					Namespace: {{ .Resource.Kind }}Namespace,
				},
				Spec: {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}Spec{
					Size: 1,
				},
			}
			Expect(k8sClient.Create(ctx, {{ lower .Resource.Kind }})).Should(Succeed())

			{{ .Resource.Kind }}LookupKey := types.NamespacedName{Name: {{ .Resource.Kind }}Name, Namespace: {{ .Resource.Kind }}Namespace}
			created{{ .Resource.Kind }} := &{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, {{ .Resource.Kind }}LookupKey, created{{ .Resource.Kind }})
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			
			By("By checking the {{ .Resource.Kind }} have one active deployment")
			deployment := &appsv1.Deployment{}
			Eventually(
				getResourceFunc(ctx, client.ObjectKey{Name: {{ .Resource.Kind }}Name, Namespace: {{ .Resource.Kind }}Namespace}, deployment),
				duration, interval).Should(BeNil())
		})
	})

})
`
