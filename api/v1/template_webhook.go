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
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"unicode"

	"github.com/disaster37/monitoring-operator/api/shared"
	sprig "github.com/go-task/slim-sprig"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"
)

// SetupWebhookWithManager will setup the manager to manage the webhooks
func SetupTemplateWebhookWithManager(mgr ctrl.Manager, client client.Client) error {
	shared.Client = client

	return ctrl.NewWebhookManagedBy(mgr).
		For(&Template{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-monitor-k8s-webcenter-fr-v1-template,mutating=false,failurePolicy=fail,sideEffects=None,groups=monitor.k8s.webcenter.fr,resources=templates,verbs=create;update,versions=v1,name=template.monitor.k8s.webcenter.fr,admissionReviewVersions=v1

var _ webhook.Validator = &Template{}

func (r *Template) validateTemplate() *field.Error {
	placeholders := map[string]any{
		"templateName":      r.Name,
		"templateNamespace": r.Namespace,
		"name":              "test",
		"namespace":         "default",
		"labels":            map[string]any{},
		"annotations":       map[string]any{},
	}

	// Check the yaml template is valid
	templateParser := template.New("template").Funcs(sprig.FuncMap())
	if r.Spec.TemplateDelimiter != nil {
		templateParser.Delims(r.Spec.TemplateDelimiter.Left, r.Spec.TemplateDelimiter.Right)
	}

	tGen, err := templateParser.Parse(r.Spec.Template)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("template"), r.Spec.Template, fmt.Sprintf("Error when parse template with golang template: %s", err.Error()))
	}
	buf := bytes.NewBufferString("")
	if err = tGen.Execute(buf, placeholders); err != nil {
		return field.Invalid(field.NewPath("spec").Child("template"), r.Spec.Template, fmt.Sprintf("Error when execute template with golang template: %s", err.Error()))
	}

	cleanTemplate := strings.TrimFunc(buf.String(), func(r rune) bool {
		return unicode.IsSpace(r)
	})
	if cleanTemplate == "" || cleanTemplate == "---" {
		return nil
	}

	data := map[string]any{}
	if err := yaml.Unmarshal(buf.Bytes(), &data); err != nil {
		return field.Invalid(field.NewPath("spec").Child("template"), r.Spec.Template, fmt.Sprintf("Error when validate yaml schema: %s", err.Error()))
	}

	if r.Spec.Type == "" {
		if data["apiVersion"] == nil || data["kind"] == nil {
			return field.Invalid(field.NewPath("spec").Child("template"), r.Spec.Template, fmt.Sprintf("You need to provide the 'apiVersion' and 'kind' on given template: '%s'", cleanTemplate))
		}
	}

	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Template) ValidateCreate() (admission.Warnings, error) {
	shared.Logger.Debugf("validate create %s/%s", r.Namespace, r.Name)
	var allErrs field.ErrorList

	if err := r.validateTemplate(); err != nil {
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
func (r *Template) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	shared.Logger.Debugf("validate create %s/%s", r.Namespace, r.Name)
	var allErrs field.ErrorList

	if err := r.validateTemplate(); err != nil {
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
func (r *Template) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}
