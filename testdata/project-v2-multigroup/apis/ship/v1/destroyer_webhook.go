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

package v1

import (
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var destroyerlog = logf.Log.WithName("destroyer-resource")

func (r *Destroyer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:webhookVersions={v1beta1},path=/mutate-ship-testproject-org-v1-destroyer,mutating=true,failurePolicy=fail,groups=ship.testproject.org,resources=destroyers,verbs=create;update,versions=v1,name=mdestroyer.kb.io,sideEffects=None,admissionReviewVersions={v1beta1}
var _ webhook.Defaulter = &Destroyer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Destroyer) Default() {
	destroyerlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}
