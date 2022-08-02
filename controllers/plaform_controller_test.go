package controllers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/disaster37/operator-sdk-extra/pkg/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *ControllerTestSuite) TestPlatformController() {
	key := types.NamespacedName{
		Name:      "t-platform-" + helpers.RandomString(10),
		Namespace: "default",
	}
	platform := &v1alpha1.Platform{}
	data := map[string]any{
		"platforms":     t.platforms,
		"dinamicClient": dynamic.NewForConfigOrDie(t.cfg),
		"k8sClient":     kubernetes.NewForConfigOrDie(t.cfg),
	}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, platform, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreatePlatformStep(),
		doListPlatformStep(),
		doUpdatePlatformStep(),
		doUpdatePlatformSecretStep(),
		doDeletePlatformStep(),
	}

	testCase.Run()

}

func doCreatePlatformStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new Plaform %s/%s ===", key.Namespace, key.Name)

			//	Create secret that store platform credentials
			secret := &core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Data: map[string][]byte{
					"username": []byte("admin"),
					"password": []byte("pass"),
				},
			}
			if err = c.Create(context.Background(), secret); err != nil {
				return err
			}

			// Create platform
			platform := &v1alpha1.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: v1alpha1.PlatformSpec{
					IsDefault:    false,
					PlatformType: "centreon",
					CentreonSettings: &v1alpha1.PlatformSpecCentreonSettings{
						URL:                   "http://localhost",
						SelfSignedCertificate: true,
						Secret:                key.Name,
					},
				},
			}

			if err = c.Create(context.Background(), platform); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			p := &v1alpha1.Platform{}
			var d any

			d, err = helper.Get(data, "platforms")
			if err != nil {
				t.Fatal(err)
			}
			platforms := d.(map[string]*ComputedPlatform)

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, p); err != nil {
					t.Fatal("Plaform not found")
				}

				if condition.IsStatusConditionPresentAndEqual(p.Status.Conditions, PlatformCondition, metav1.ConditionTrue) {
					return nil
				}
				return errors.New("Platform not yet loaded")
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Platform: %s", err.Error())
			}

			assert.NotEmpty(t, platforms[key.Name])
			assert.NotNil(t, platforms[key.Name].client)
			assert.NotEmpty(t, platforms[key.Name].hash)

			data["platform"] = platforms[key.Name]
			return nil
		},
	}
}

func doUpdatePlatformStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update Plaform %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Plaform is null")
			}
			p := o.(*v1alpha1.Platform)

			data["version"] = p.ResourceVersion

			p.Spec.CentreonSettings.URL = "http://localhost2"

			if err = c.Update(context.Background(), p); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			p := &v1alpha1.Platform{}
			var d any

			d, err = helper.Get(data, "platforms")
			if err != nil {
				t.Fatal(err)
			}
			platforms := d.(map[string]*ComputedPlatform)

			d, err = helper.Get(data, "platform")
			if err != nil {
				t.Fatal(err)
			}
			platform := d.(*ComputedPlatform)

			d, err = helper.Get(data, "version")
			if err != nil {
				t.Fatal(err)
			}
			version := d.(string)

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, p); err != nil {
					t.Fatalf("Error when get Centreon service: %s", err.Error())
				}

				if p.ResourceVersion == version {
					return errors.New("Not yet updated")
				}

				return nil

			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get platform: %s", err.Error())
			}

			assert.Equal(t, "http://localhost2", platforms[key.Name].platform.Spec.CentreonSettings.URL)
			assert.NotEqual(t, platforms[key.Name].hash, platform.hash)
			assert.NotEqual(t, platforms[key.Name].client, platform.client)
			return nil
		},
	}
}

func doUpdatePlatformSecretStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update Plaform secret %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Plaform is null")
			}

			secret := &core.Secret{}
			if err = c.Get(context.Background(), key, secret); err != nil {
				return err
			}
			secret.Data["username"] = []byte("admin2")

			if err = c.Update(context.Background(), secret); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			p := &v1alpha1.Platform{}
			var d any

			d, err = helper.Get(data, "platforms")
			if err != nil {
				t.Fatal(err)
			}
			platforms := d.(map[string]*ComputedPlatform)

			d, err = helper.Get(data, "platform")
			if err != nil {
				t.Fatal(err)
			}
			platform := d.(*ComputedPlatform)

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, p); err != nil {
					t.Fatalf("Error when get Centreon service: %s", err.Error())
				}

				if platforms[key.Name].client == platform.client {
					return errors.New("Not yet updated")
				}

				return nil

			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get platform: %s", err.Error())
			}

			assert.Equal(t, "http://localhost2", platforms[key.Name].platform.Spec.CentreonSettings.URL)
			assert.NotEqual(t, platforms[key.Name].client, platform.client)
			return nil
		},
	}
}

func doListPlatformStep() test.TestStep {
	return test.TestStep{
		Name: "list",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Info("=== List Plaforms ===")

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			var d any

			d, err = helper.Get(data, "dinamicClient")
			if err != nil {
				t.Fatal(err)
			}
			dinamicClient := d.(dynamic.Interface)

			d, err = helper.Get(data, "k8sClient")
			if err != nil {
				t.Fatal(err)
			}
			k8sClient := d.(kubernetes.Interface)

			log := logrus.NewEntry(logrus.New())
			log.Logger.SetLevel(logrus.DebugLevel)

			platforms, err := ComputedPlatformList(context.Background(), dinamicClient, k8sClient, log)
			if err != nil {
				t.Fatal(err)
			}

			assert.NotNil(t, platforms[key.Name])

			return nil
		},
	}
}

func doDeletePlatformStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete Plaform %s/%s ===", key.Namespace, key.Name)
			if o == nil {
				return errors.New("Plaform is null")
			}
			p := o.(*v1alpha1.Platform)

			wait := int64(0)
			if err = c.Delete(context.Background(), p, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			p := &v1alpha1.Platform{}
			isDeleted := false

			var d any

			d, err = helper.Get(data, "platforms")
			if err != nil {
				t.Fatal(err)
			}
			platforms := d.(map[string]*ComputedPlatform)

			// Object can be deleted or marked as deleted
			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, p); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return nil

			}, time.Second*30, time.Second*1)

			if err != nil || isTimeout {
				t.Fatalf("Platform not deleted: %s", err.Error())
			}
			assert.True(t, isDeleted)
			assert.Nil(t, platforms[key.Name])

			return nil
		},
	}
}
