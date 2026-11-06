package template

import (
	"context"
	"os"
	"testing"

	"emperror.dev/errors"
	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/operator-sdk-extra/pkg/apis/shared"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const testParentLabel = "monitor.k8s.webcenter.fr/parent"

type sentinelObjectFactory struct {
	name          string
	finalizer     shared.FinalizerName
	clusterScoped bool
	newObject     func(name, namespace string) client.Object
}

func sentinelObjectFactories() []sentinelObjectFactory {
	return []sentinelObjectFactory{
		{
			name:          "namespace",
			finalizer:     NamespaceFinalizer,
			clusterScoped: true,
			newObject: func(name, _ string) client.Object {
				return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
			},
		},
		{
			name:      "ingress",
			finalizer: IngressFinalizer,
			newObject: func(name, namespace string) client.Object {
				return &networkv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
			},
		},
		{
			name:      "route",
			finalizer: RouteFinalizer,
			newObject: func(name, namespace string) client.Object {
				return &routev1.Route{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
			},
		},
		{
			name:      "certificate",
			finalizer: CertificateFinalizer,
			newObject: func(name, namespace string) client.Object {
				return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
			},
		},
		{
			name:          "node",
			finalizer:     NodeFinalizer,
			clusterScoped: true,
			newObject: func(name, _ string) client.Object {
				return &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: name}}
			},
		},
	}
}

func newTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()

	s := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(s))
	require.NoError(t, networkv1.AddToScheme(s))
	require.NoError(t, routev1.AddToScheme(s))
	require.NoError(t, centreoncrd.AddToScheme(s))

	return s
}

func newTestReconciler(t *testing.T, c client.Client, finalizer shared.FinalizerName) controller.SentinelReconciler {
	t.Helper()

	return NewSentinelReconciler(
		c,
		"test",
		finalizer,
		logrus.NewEntry(logrus.StandardLogger()),
		record.NewFakeRecorder(100),
	)
}

func newTestAction(c client.Client) controller.SentinelReconcilerAction {
	return NewTemplateReconciler(c, record.NewFakeRecorder(100))
}

func childNamespace(factory sentinelObjectFactory, name, namespace string) string {
	switch factory.name {
	case "namespace":
		return name
	case "node":
		return "default"
	default:
		return namespace
	}
}

func parentKey(factory sentinelObjectFactory, name, namespace string) types.NamespacedName {
	if factory.clusterScoped {
		return types.NamespacedName{Name: name}
	}
	return types.NamespacedName{Name: name, Namespace: namespace}
}

func newChildService(name, namespace, parent string) *centreoncrd.CentreonService {
	return &centreoncrd.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				testParentLabel: parent,
			},
		},
	}
}

func newChildServiceGroup(name, namespace, parent string) *centreoncrd.CentreonServiceGroup {
	return &centreoncrd.CentreonServiceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				testParentLabel: parent,
			},
		},
	}
}

func TestReconcile_AddFinalizerWhenTemplateAnnotationPresent(t *testing.T) {
	for _, factory := range sentinelObjectFactories() {
		t.Run(factory.name, func(t *testing.T) {
			if factory.name == "node" {
				t.Setenv("POD_NAMESPACE", "default")
			}
			s := newTestScheme(t)
			namespace := "default"
			name := "parent"
			obj := factory.newObject(name, namespace)
			obj.SetAnnotations(map[string]string{
				"monitor.k8s.webcenter.fr/templates": "[]",
			})

			c := fake.NewClientBuilder().WithScheme(s).WithObjects(obj).Build()
			r := newTestReconciler(t, c, factory.finalizer)

			res, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: parentKey(factory, name, namespace)}, factory.newObject(name, namespace), map[string]any{}, newTestAction(c))
			require.NoError(t, err)
			assert.True(t, res.Requeue)

			updated := factory.newObject(name, namespace)
			require.NoError(t, c.Get(context.Background(), parentKey(factory, name, namespace), updated))
			assert.Contains(t, updated.GetFinalizers(), factory.finalizer.String())
		})
	}
}

func TestReconcile_NoFinalizerWhenNoTemplateAnnotation(t *testing.T) {
	for _, factory := range sentinelObjectFactories() {
		t.Run(factory.name, func(t *testing.T) {
			if factory.name == "node" {
				t.Setenv("POD_NAMESPACE", "default")
			}
			s := newTestScheme(t)
			namespace := "default"
			name := "parent"
			obj := factory.newObject(name, namespace)

			c := fake.NewClientBuilder().WithScheme(s).WithObjects(obj).Build()
			r := newTestReconciler(t, c, factory.finalizer)

			res, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: parentKey(factory, name, namespace)}, factory.newObject(name, namespace), map[string]any{}, newTestAction(c))
			require.NoError(t, err)
			assert.Equal(t, ctrl.Result{}, res)

			updated := factory.newObject(name, namespace)
			require.NoError(t, c.Get(context.Background(), parentKey(factory, name, namespace), updated))
			assert.Empty(t, updated.GetFinalizers())
		})
	}
}

func TestReconcile_IgnoreReconcileOnActiveObject(t *testing.T) {
	s := newTestScheme(t)
	obj := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "parent",
			Namespace: "default",
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/templates":              "[]",
				"operator-sdk-extra.webcenter.fr/ignoreReconcile": "true",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(s).WithObjects(obj).Build()
	r := newTestReconciler(t, c, CertificateFinalizer)

	res, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: "parent", Namespace: "default"}}, &corev1.Secret{}, map[string]any{}, newTestAction(c))
	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	updated := &corev1.Secret{}
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Name: "parent", Namespace: "default"}, updated))
	assert.Empty(t, updated.GetFinalizers())
}

func TestReconcile_DeleteParentCleansChildrenAndRemovesFinalizer(t *testing.T) {
	for _, factory := range sentinelObjectFactories() {
		t.Run(factory.name, func(t *testing.T) {
			if factory.name == "node" {
				t.Setenv("POD_NAMESPACE", "default")
			}
			s := newTestScheme(t)
			namespace := "default"
			name := "parent"
			parent := factory.newObject(name, namespace)
			parent.SetFinalizers([]string{factory.finalizer.String()})
			parent.SetAnnotations(map[string]string{
				"monitor.k8s.webcenter.fr/templates": "[]",
			})

			childNS := childNamespace(factory, name, namespace)
			parentLabel := childNS + "." + name
			service := newChildService("child-service", childNS, parentLabel)
			serviceGroup := newChildServiceGroup("child-service-group", childNS, parentLabel)

			c := fake.NewClientBuilder().WithScheme(s).WithObjects(parent, service, serviceGroup).Build()
			r := newTestReconciler(t, c, factory.finalizer)

			require.NoError(t, c.Delete(context.Background(), factory.newObject(name, namespace)))

			res, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: parentKey(factory, name, namespace)}, factory.newObject(name, namespace), map[string]any{}, newTestAction(c))
			require.NoError(t, err)
			assert.Equal(t, ctrl.Result{}, res)

			assert.True(t, k8serrors.IsNotFound(c.Get(context.Background(), types.NamespacedName{Name: "child-service", Namespace: childNS}, &centreoncrd.CentreonService{})))
			assert.True(t, k8serrors.IsNotFound(c.Get(context.Background(), types.NamespacedName{Name: "child-service-group", Namespace: childNS}, &centreoncrd.CentreonServiceGroup{})))
			assert.True(t, k8serrors.IsNotFound(c.Get(context.Background(), parentKey(factory, name, namespace), factory.newObject(name, namespace))))
		})
	}
}

func TestReconcile_DeleteParentWithIgnoreReconcileStillCleansUp(t *testing.T) {
	s := newTestScheme(t)
	parent := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "parent",
			Namespace:  "default",
			Finalizers: []string{CertificateFinalizer.String()},
			Annotations: map[string]string{
				"operator-sdk-extra.webcenter.fr/ignoreReconcile": "true",
			},
		},
	}
	service := newChildService("child-service", "default", "default.parent")

	c := fake.NewClientBuilder().WithScheme(s).WithObjects(parent, service).Build()
	r := newTestReconciler(t, c, CertificateFinalizer)

	require.NoError(t, c.Delete(context.Background(), &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "parent", Namespace: "default"}}))

	res, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: "parent", Namespace: "default"}}, &corev1.Secret{}, map[string]any{}, newTestAction(c))
	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	assert.True(t, k8serrors.IsNotFound(c.Get(context.Background(), types.NamespacedName{Name: "child-service", Namespace: "default"}, &centreoncrd.CentreonService{})))
	assert.True(t, k8serrors.IsNotFound(c.Get(context.Background(), types.NamespacedName{Name: "parent", Namespace: "default"}, &corev1.Secret{})))
}

func TestReconcile_LegacyFinalizerRemoval(t *testing.T) {
	s := newTestScheme(t)
	parent := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "parent",
			Namespace:  "default",
			Finalizers: []string{TemplateFinalizer.String()},
		},
	}
	service := newChildService("child-service", "default", "default.parent")

	c := fake.NewClientBuilder().WithScheme(s).WithObjects(parent, service).Build()
	r := newTestReconciler(t, c, CertificateFinalizer)

	require.NoError(t, c.Delete(context.Background(), &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "parent", Namespace: "default"}}))

	res, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: "parent", Namespace: "default"}}, &corev1.Secret{}, map[string]any{}, newTestAction(c))
	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	assert.True(t, k8serrors.IsNotFound(c.Get(context.Background(), types.NamespacedName{Name: "child-service", Namespace: "default"}, &centreoncrd.CentreonService{})))
	assert.True(t, k8serrors.IsNotFound(c.Get(context.Background(), types.NamespacedName{Name: "parent", Namespace: "default"}, &corev1.Secret{})))
}

func TestReconcile_AnnotationRemovedFromActiveResource(t *testing.T) {
	s := newTestScheme(t)
	parent := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "parent",
			Namespace:  "default",
			Finalizers: []string{CertificateFinalizer.String()},
		},
	}
	service := newChildService("child-service", "default", "default.parent")

	c := fake.NewClientBuilder().WithScheme(s).WithObjects(parent, service).Build()
	r := newTestReconciler(t, c, CertificateFinalizer)

	res, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: "parent", Namespace: "default"}}, &corev1.Secret{}, map[string]any{}, newTestAction(c))
	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	assert.True(t, k8serrors.IsNotFound(c.Get(context.Background(), types.NamespacedName{Name: "child-service", Namespace: "default"}, &centreoncrd.CentreonService{})))

	updated := &corev1.Secret{}
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Name: "parent", Namespace: "default"}, updated))
	assert.Empty(t, updated.GetFinalizers())
}

func TestReconcile_DeleteParentWithBrokenTemplate(t *testing.T) {
	s := newTestScheme(t)
	parent := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "parent",
			Namespace:  "default",
			Finalizers: []string{CertificateFinalizer.String()},
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/templates": "{not-json",
			},
		},
	}
	service := newChildService("child-service", "default", "default.parent")

	c := fake.NewClientBuilder().WithScheme(s).WithObjects(parent, service).Build()
	r := newTestReconciler(t, c, CertificateFinalizer)

	require.NoError(t, c.Delete(context.Background(), &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "parent", Namespace: "default"}}))

	res, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: "parent", Namespace: "default"}}, &corev1.Secret{}, map[string]any{}, newTestAction(c))
	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	assert.True(t, k8serrors.IsNotFound(c.Get(context.Background(), types.NamespacedName{Name: "child-service", Namespace: "default"}, &centreoncrd.CentreonService{})))
	assert.True(t, k8serrors.IsNotFound(c.Get(context.Background(), types.NamespacedName{Name: "parent", Namespace: "default"}, &corev1.Secret{})))
}

func TestReconcile_DeleteParentChildAlreadyNotFound(t *testing.T) {
	s := newTestScheme(t)
	parent := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "parent",
			Namespace:  "default",
			Finalizers: []string{CertificateFinalizer.String()},
		},
	}

	c := fake.NewClientBuilder().WithScheme(s).WithObjects(parent).Build()
	r := newTestReconciler(t, c, CertificateFinalizer)

	require.NoError(t, c.Delete(context.Background(), &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "parent", Namespace: "default"}}))

	res, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: "parent", Namespace: "default"}}, &corev1.Secret{}, map[string]any{}, newTestAction(c))
	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	assert.True(t, k8serrors.IsNotFound(c.Get(context.Background(), types.NamespacedName{Name: "parent", Namespace: "default"}, &corev1.Secret{})))
}

func TestReconcile_DeleteParentTransientDeleteErrorKeepsFinalizer(t *testing.T) {
	s := newTestScheme(t)
	parent := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "parent",
			Namespace:  "default",
			Finalizers: []string{CertificateFinalizer.String()},
		},
	}
	service := newChildService("child-service", "default", "default.parent")

	c := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(parent, service).
		WithInterceptorFuncs(interceptor.Funcs{
			Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				if _, ok := obj.(*centreoncrd.CentreonService); ok {
					return errors.New("transient delete error")
				}
				return c.Delete(ctx, obj, opts...)
			},
		}).
		Build()
	r := newTestReconciler(t, c, CertificateFinalizer)

	require.NoError(t, c.Delete(context.Background(), &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "parent", Namespace: "default"}}))

	_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: "parent", Namespace: "default"}}, &corev1.Secret{}, map[string]any{}, newTestAction(c))
	require.Error(t, err)

	updated := &corev1.Secret{}
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Name: "parent", Namespace: "default"}, updated))
	assert.Contains(t, updated.GetFinalizers(), CertificateFinalizer.String())
}

func TestReconcile_NotFoundParent(t *testing.T) {
	s := newTestScheme(t)
	c := fake.NewClientBuilder().WithScheme(s).Build()
	r := newTestReconciler(t, c, CertificateFinalizer)

	res, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: "missing", Namespace: "default"}}, &corev1.Secret{}, map[string]any{}, newTestAction(c))
	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)
}

func TestResolveChildNamespace(t *testing.T) {
	t.Run("namespace", func(t *testing.T) {
		ns, err := resolveChildNamespace(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "my-ns"}})
		require.NoError(t, err)
		assert.Equal(t, "my-ns", ns)
	})

	t.Run("node", func(t *testing.T) {
		t.Setenv("POD_NAMESPACE", "operator-ns")
		ns, err := resolveChildNamespace(&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "my-node"}})
		require.NoError(t, err)
		assert.Equal(t, "operator-ns", ns)
	})

	t.Run("node without operator namespace", func(t *testing.T) {
		t.Setenv("POD_NAMESPACE", "operator-ns")
		require.NoError(t, os.Unsetenv("POD_NAMESPACE"))
		_, err := resolveChildNamespace(&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "my-node"}})
		require.Error(t, err)
	})

	t.Run("namespaced resources", func(t *testing.T) {
		objects := []client.Object{
			&networkv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ingress", Namespace: "my-ns"}},
			&routev1.Route{ObjectMeta: metav1.ObjectMeta{Name: "route", Namespace: "my-ns"}},
			&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: "my-ns"}},
		}
		for _, obj := range objects {
			ns, err := resolveChildNamespace(obj)
			require.NoError(t, err)
			assert.Equal(t, "my-ns", ns)
		}
	})
}
