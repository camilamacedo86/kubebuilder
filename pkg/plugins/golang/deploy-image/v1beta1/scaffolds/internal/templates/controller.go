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

package templates

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Controller{}

// Controller scaffolds the file that defines the controller for a CRD or a builtin resource
// nolint:maligned
type Controller struct {
	machinery.TemplateMixin
	machinery.MultiGroupMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin

	Image string
}

// SetTemplateDefaults implements file.Template
func (f *Controller) SetTemplateDefaults() error {
	if f.Path == "" {
		if f.MultiGroup {
			f.Path = filepath.Join("controllers", "%[group]", "%[kind]_controller.go")
		} else {
			f.Path = filepath.Join("controllers", "%[kind]_controller.go")
		}
	}
	f.Path = f.Resource.Replacer().Replace(f.Path)

	f.TemplateBody = controllerTemplate

	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

//nolint:lll
const controllerTemplate = `{{ .Boilerplate }}

package controllers

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/addon"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/addon/pkg/status"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative"

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
)

var _ reconcile.Reconciler = &{{ .Resource.Kind }}Reconciler{}

// {{ .Resource.Kind }}Reconciler reconciles a {{ .Resource.Kind }} object
type {{ .Resource.Kind }}Reconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	declarative.Reconciler
}

//+kubebuilder:rbac:groups={{ .Resource.QualifiedGroup }},resources={{ .Resource.Plural }},verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups={{ .Resource.QualifiedGroup }},resources={{ .Resource.Plural }}/status,verbs=get;update;patch
//+kubebuilder:rbac:groups={{ .Resource.QualifiedGroup }},resources={{ .Resource.Plural }}/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

func (r *{{ .Resource.Kind }}Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("{{ .Resource.Kind }}", req.NamespacedName)

	// Fetch the {{ .Resource.Kind }} instance
	{{ .Resource.Kind }} := &{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}{}
	err := r.Get(ctx, req.NamespacedName, {{ .Resource.Kind }})
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("{{ .Resource.Kind }} resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get {{ .Resource.Kind }}")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: {{ .Resource.Kind }}.Name, Namespace: {{ .Resource.Kind }}.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentFor{{ .Resource.Kind }}({{ .Resource.Kind }})
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := {{ .Resource.Kind }}.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}
	
    // TODO: add here code implementation to update/manage the status
    
	return ctrl.Result{}, nil
}

// deploymentFor{{ .Resource.Kind }} returns a {{ .Resource.Kind }} Deployment object
func (r *{{ .Resource.Kind }}Reconciler) deploymentFor{{ .Resource.Kind }}(m *{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}) *appsv1.Deployment {
	ls := labelsFor{{ .Resource.Kind }}(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   imageFor{{ .Resource.Kind }}(m.Name),
						Name:    {{ .Resource.Kind }},
                        ImagePullPolicy: {{ .Resource.Kind }}.Spec.ContainerImagePullPolicy,
						Command: []string{"{{ .Resource.Kind }}"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: {{ .Resource.Kind }}.Spec.ContainerPort,
							Name:          "{{ .Resource.Kind }}",
						}},
					}},
				},
			},
		},
	}
	// Set {{ .Resource.Kind }} instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// labelsFor{{ .Resource.Kind }} returns the labels for selecting the resources
// belonging to the given {{ .Resource.Kind }} CR name.
func labelsFor{{ .Resource.Kind }}(name string) map[string]string {
	return map[string]string{"type": "{{ .Resource.Kind }}", "{{ .Resource.Kind }}_cr": name}
}

// imageFor{{ .Resource.Kind }} returns the image for the resources
// belonging to the given {{ .Resource.Kind }} CR name.
func imageFor{{ .Resource.Kind }}(name string) string {
	// TODO: this method will return the value of the envvar create to store the image:tag informed
}

func (r *{{ .Resource.Kind }}Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.{{ .Resource.Kind }}{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
// .Image adding this for just checking
`
