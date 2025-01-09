/*
Copyright 2021.

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
	"context"
	"fmt"
	"strings"

	"github.com/disaster37/monitoring-operator/api/shared"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupWebhookWithManager will setup the manager to manage the webhooks
func SetupPlatformWebhookWithManager(mgr ctrl.Manager, client client.Client) error {
	shared.Client = client

	return ctrl.NewWebhookManagedBy(mgr).
		For(&Platform{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-monitor-k8s-webcenter-fr-v1-platform,mutating=false,failurePolicy=fail,sideEffects=None,groups=monitor.k8s.webcenter.fr,resources=platforms,verbs=create;update,versions=v1,name=platform.monitor.k8s.webcenter.fr,admissionReviewVersions=v1

var _ webhook.Validator = &Platform{}

func (r *Platform) validateRequiredFields() *field.Error {
	switch r.Spec.PlatformType {
	case "centreon":
		if r.Spec.CentreonSettings == nil {
			return field.Required(field.NewPath("spec").Child("centreonSettings"), "You need to provide the Centreon settings")
		}
	}

	return nil
}

func (r *Platform) validateNamespace() *field.Error {
	ns, err := helpers.GetOperatorNamespace()
	if err != nil {
		panic(err)
	}
	if r.Namespace != ns {
		return field.Forbidden(field.NewPath("metadata").Child("namespace"), "You can only create Platform resource on same namespace operator")
	}

	return nil
}

func (r *Platform) validateUnicityOfDefaultPlatform() *field.Error {
	// Check that only one platform as default
	listObjects := &PlatformList{}
	fs := fields.ParseSelectorOrDie(fmt.Sprintf("spec.isDefault=%s", helpers.BoolToString(&r.Spec.IsDefault)))
	if err := shared.Client.List(context.Background(), listObjects, &client.ListOptions{FieldSelector: fs}); err != nil {
		panic(err)
	}
	if len(listObjects.Items) > 0 {
		isError := false
		existingResources := make([]string, 0, len(listObjects.Items))
		for _, ag := range listObjects.Items {
			// exclude themself
			if ag.UID != r.UID {
				existingResources = append(existingResources, fmt.Sprintf("'%s/%s'", ag.Namespace, ag.Name))
				isError = true
			}
		}
		if isError {
			return field.Duplicate(field.NewPath("spec").Child("isDefault"), fmt.Sprintf("There are some other platform set as default platform: %s", strings.Join(existingResources, ", ")))
		}
	}
	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Platform) ValidateCreate() (admission.Warnings, error) {
	shared.Logger.Debugf("validate create %s/%s", r.Namespace, r.Name)
	var allErrs field.ErrorList

	if err := r.validateNamespace(); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := r.validateRequiredFields(); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := r.validateUnicityOfDefaultPlatform(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) > 0 {
		return nil, apierrors.NewInvalid(
			r.GroupVersionKind().GroupKind(),
			r.Name, allErrs)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Platform) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	shared.Logger.Debugf("validate create %s/%s", r.Namespace, r.Name)
	var allErrs field.ErrorList

	if err := r.validateNamespace(); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := r.validateRequiredFields(); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := r.validateUnicityOfDefaultPlatform(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) > 0 {
		return nil, apierrors.NewInvalid(
			r.GroupVersionKind().GroupKind(),
			r.Name, allErrs)
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Platform) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}
