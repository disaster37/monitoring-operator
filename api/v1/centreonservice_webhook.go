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
func SetupCentreonServiceWebhookWithManager(mgr ctrl.Manager, client client.Client) error {
	shared.Client = client

	return ctrl.NewWebhookManagedBy(mgr).
		For(&CentreonService{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-monitor-k8s-webcenter-fr-v1-centreonservice,mutating=false,failurePolicy=fail,sideEffects=None,groups=monitor.k8s.webcenter.fr,resources=centreonservices,verbs=create;update,versions=v1,name=centreonservice.monitor.k8s.webcenter.fr,admissionReviewVersions=v1

var _ webhook.Validator = &CentreonService{}

// validateField permit to validate the centreonService fields
func (r *CentreonService) validateField() *field.Error {
	if r.Spec.CheckCommand == "" && r.Spec.Template == "" {
		return field.Required(field.NewPath("spec"), "You need to provide 'spec.checkCommand' or 'spec.template' field")
	}

	return nil
}

func (r *CentreonService) validateImmatablePlatform(current, old *CentreonService) *field.Error {
	if current.GetPlatform() != old.GetPlatform() {
		return field.Forbidden(field.NewPath("spec").Child("platformRef"), "The field 'spec.platformRef' is immutable")
	}
	return nil
}

func (r *CentreonService) validateResourceUnicity() *field.Error {
	// Check if resource already exist with same name on some remote target platform
	listObjects := &CentreonServiceList{}
	fs := fields.ParseSelectorOrDie(fmt.Sprintf("spec.externalName=%s/%s,spec.targetPlatform=%s", r.Spec.Host, r.GetExternalName(), r.GetPlatform()))
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
			return field.Duplicate(field.NewPath("spec").Child("name"), fmt.Sprintf("There are some same resource that already target the same monitoring platform with the same name: %s", strings.Join(existingResources, ", ")))
		}
	}

	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CentreonService) ValidateCreate() (admission.Warnings, error) {
	shared.Logger.Debugf("validate create %s/%s", r.Namespace, r.Name)
	var allErrs field.ErrorList

	if err := r.validateField(); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := r.validateResourceUnicity(); err != nil {
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
func (r *CentreonService) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	shared.Logger.Debugf("validate update %s/%s", r.Namespace, r.Name)
	var allErrs field.ErrorList
	oldCS := old.(*CentreonService)

	if err := r.validateImmatablePlatform(r, oldCS); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := r.validateField(); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := r.validateResourceUnicity(); err != nil {
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
func (r *CentreonService) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}
