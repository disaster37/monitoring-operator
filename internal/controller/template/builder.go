package template

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"

	"dario.cat/mergo"
	"emperror.dev/errors"
	"github.com/disaster37/monitoring-operator/api/shared"
	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	sprig "github.com/go-task/slim-sprig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type builder struct {
	placehodlers                 map[string]any
	supportedTemplateObjects     map[string]client.Object
	supportedTemplateObjectsList []shared.ObjectList
	sourceObject                 client.Object
	scheme                       runtime.ObjectTyper
}

func newBuilder(o client.Object, scheme runtime.ObjectTyper) *builder {
	return &builder{
		sourceObject:                 o,
		supportedTemplateObjects:     make(map[string]client.Object),
		supportedTemplateObjectsList: make([]shared.ObjectList, 0),
		placehodlers: map[string]any{
			"name":        o.GetName(),
			"namespace":   o.GetNamespace(),
			"labels":      o.GetLabels(),
			"annotations": o.GetAnnotations(),
		},
		scheme: scheme,
	}
}

func (h *builder) AddPlaceholders(placeholders map[string]any) *builder {
	if err := mergo.Merge(h.placehodlers, placeholders, mergo.WithAppendSlice); err != nil {
		panic(err)
	}

	return h
}

func (h *builder) For(o client.Object, oList shared.ObjectList) *builder {
	o = helpers.GetObjectWithMeta(o, h.scheme)
	h.supportedTemplateObjects[helpers.GetObjectType(o)] = o
	h.supportedTemplateObjectsList = append(h.supportedTemplateObjectsList, oList)
	return h
}

func (h *builder) Lists() []shared.ObjectList {
	lists := make([]shared.ObjectList, 0, len(h.supportedTemplateObjectsList))
	for _, oList := range h.supportedTemplateObjectsList {
		lists = append(lists, reflect.New(reflect.TypeOf(oList)).Interface().(shared.ObjectList))
	}
	return lists
}

func (h *builder) Objects() []client.Object {
	list := make([]client.Object, 0, len(h.supportedTemplateObjects))
	for _, o := range h.supportedTemplateObjects {
		list = append(list, reflect.New(reflect.TypeOf(o)).Interface().(client.Object))
	}

	return list

}

func (h *builder) Process(t *centreoncrd.Template) (object client.Object, err error) {
	meta := &metav1.TypeMeta{}
	object = helpers.GetObjectWithMeta(object, h.scheme)

	h.placehodlers["templateName"] = t.Name
	h.placehodlers["templateNamespace"] = t.Namespace

	tGen, err := template.New("template").Funcs(sprig.FuncMap()).Parse(t.Spec.Template)
	if err != nil {
		return nil, errors.Wrapf(err, "Error when parse template %s/%s from %s/%s", t.Namespace, t.Name, h.sourceObject.GetNamespace(), h.sourceObject.GetName())
	}
	buf := bytes.NewBufferString("")
	if err = tGen.Execute(buf, h.placehodlers); err != nil {
		return nil, errors.Wrapf(err, "Error when execute template %s/%s from %s/%s", t.Namespace, t.Name, h.sourceObject.GetNamespace(), h.sourceObject.GetName())
	}

	// We need to support old stategy when type is provided instead to set the full object on template
	if t.Spec.Type != "" {

		// Process resource name
		targetResourceName, err := processName(t, h.placehodlers)
		if err != nil {
			return nil, errors.Wrap(err, "Error when process template name")
		}

		switch t.Spec.Type {
		case "CentreonService":
			centreonServiceSpec := &centreoncrd.CentreonServiceSpec{}
			// Compute expected resource spec
			if err = yaml.Unmarshal(buf.Bytes(), centreonServiceSpec); err != nil {
				return nil, errors.Wrap(err, "Error when unmarshall template")
			}

			cs := &centreoncrd.CentreonService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      targetResourceName,
					Namespace: h.sourceObject.GetNamespace(),
				},
				Spec: *centreonServiceSpec,
			}

			// Check CentreonService is valid
			if !cs.IsValid() {
				return nil, fmt.Errorf("generated CentreonService is not valid: %+v", cs.Spec)
			}
			return cs, nil
		case "CentreonServiceGroup":
			centreonServiceGroupSpec := &centreoncrd.CentreonServiceGroupSpec{}
			// Compute expected resource spec
			if err = yaml.Unmarshal(buf.Bytes(), centreonServiceGroupSpec); err != nil {
				return nil, errors.Wrap(err, "Error when unmarshall expected spec")
			}

			centreonServiceGroup := &centreoncrd.CentreonServiceGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      targetResourceName,
					Namespace: h.sourceObject.GetNamespace(),
				},
				Spec: *centreonServiceGroupSpec,
			}

			// Check CentreonServiceGroup is valid
			if !centreonServiceGroup.IsValid() {
				return nil, fmt.Errorf("generated CentreonServiceGroup is not valid: %+v", centreonServiceGroup.Spec)
			}
			return centreonServiceGroup, nil
		default:
			return nil, errors.Errorf("Template of type %s is not supported", t.Spec.Type)
		}
	}

	if err := yaml.Unmarshal(buf.Bytes(), meta); err != nil {
		return nil, errors.Wrapf(err, "Error when Unmarshall template %s/%s from %s/%s", t.Namespace, t.Name, h.sourceObject.GetNamespace(), h.sourceObject.GetName())
	}

	o, isFound := h.supportedTemplateObjects[helpers.GetObjectType(object)]
	if !isFound {
		return nil, errors.Errorf("No type '%s' found for template %s/%s from %s/%s", helpers.GetObjectType(object), t.Namespace, t.Name, h.sourceObject.GetNamespace(), h.sourceObject.GetName())
	}

	newO := helpers.CloneObject(o)

	if err = yaml.Unmarshal(buf.Bytes(), newO); err != nil {
		return nil, errors.Wrapf(err, "Error when unmarshall resource template %s/%s from %s/%s", t.Namespace, t.Name, h.sourceObject.GetNamespace(), h.sourceObject.GetName())
	}

	return newO, nil

}

// processName permit to get the resource name generated from template
// It return the template name if name is not provided
func processName(templateO *centreoncrd.Template, placeholders map[string]any) (name string, err error) {
	if templateO.Spec.Name == "" {
		return templateO.Name, nil
	}

	t, err := template.New("template").Funcs(sprig.FuncMap()).Parse(templateO.Spec.Name)
	if err != nil {
		return "", errors.Wrapf(err, "Error when parse template name %s/%s", templateO.Namespace, templateO.Name)
	}
	buf := bytes.NewBufferString("")
	if err = t.Execute(buf, placeholders); err != nil {
		return "", errors.Wrap(err, "Error when execute template")
	}

	return buf.String(), nil
}
