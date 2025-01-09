package platform

import (
	"context"
	"errors"
	"testing"
	"time"

	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/disaster37/operator-sdk-extra/pkg/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *PlatformControllerTestSuite) TestPlatformController() {
	key := types.NamespacedName{
		Name:      "t-platform-" + helpers.RandomString(10),
		Namespace: "default",
	}
	platform := &monitorapi.Platform{}
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
			platform := &monitorapi.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: monitorapi.PlatformSpec{
					IsDefault:    false,
					PlatformType: "centreon",
					CentreonSettings: &monitorapi.PlatformSpecCentreonSettings{
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
			p := &monitorapi.Platform{}
			var d any

			d, err = helper.Get(data, "platforms")
			if err != nil {
				t.Fatal(err)
			}
			platforms := d.(map[string]*ComputedPlatform)

			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, p); err != nil {
					t.Fatal("Plaform not found")
				}

				if p.GetStatus().GetObservedGeneration() == 0 {
					return errors.New("Platform not yet loaded")
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Platform: %s", err.Error())
			}

			assert.NotEmpty(t, platforms[key.Name])
			assert.NotNil(t, platforms[key.Name].Client)
			assert.NotEmpty(t, platforms[key.Name].Hash)

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
			p := o.(*monitorapi.Platform)

			data["lastGeneration"] = p.GetStatus().GetObservedGeneration()

			p.Spec.CentreonSettings.URL = "http://localhost2"

			if err = c.Update(context.Background(), p); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			p := &monitorapi.Platform{}
			var d any
			lastGeneration := data["lastGeneration"].(int64)

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

			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, p); err != nil {
					t.Fatalf("Error when get Centreon service: %s", err.Error())
				}

				if lastGeneration == p.GetStatus().GetObservedGeneration() {
					return errors.New("Not yet updated")
				}

				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get platform: %s", err.Error())
			}

			assert.Equal(t, "http://localhost2", platforms[key.Name].Platform.Spec.CentreonSettings.URL)
			assert.NotEqual(t, platforms[key.Name].Hash, platform.Hash)
			assert.NotEqual(t, platforms[key.Name].Client, platform.Client)
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
			p := &monitorapi.Platform{}
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

			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, p); err != nil {
					t.Fatalf("Error when get Centreon service: %s", err.Error())
				}

				if platforms[key.Name].Client == platform.Client {
					return errors.New("Not yet updated")
				}

				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get platform: %s", err.Error())
			}

			assert.Equal(t, "http://localhost2", platforms[key.Name].Platform.Spec.CentreonSettings.URL)
			assert.NotEqual(t, platforms[key.Name].Client, platform.Client)
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
			log := logrus.NewEntry(logrus.New())
			log.Logger.SetLevel(logrus.DebugLevel)

			platforms, err := ComputedPlatformList(context.Background(), c, log)
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
			p := o.(*monitorapi.Platform)

			wait := int64(0)
			if err = c.Delete(context.Background(), p, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			p := &monitorapi.Platform{}
			isDeleted := false

			var d any

			d, err = helper.Get(data, "platforms")
			if err != nil {
				t.Fatal(err)
			}
			platforms := d.(map[string]*ComputedPlatform)

			// Object can be deleted or marked as deleted
			isTimeout, err := test.RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, p); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return errors.New("Not yet deleted")
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
