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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *ControllerTestSuite) TestPlatformController() {
	key := types.NamespacedName{
		Name:      "t-platform-" + helpers.RandomString(10),
		Namespace: "default",
	}
	platform := &v1alpha1.Platform{}
	data := map[string]any{
		"platforms": t.platforms,
	}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, platform, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreatePlatformStep(),
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
					Name:         "test",
					IsDefault:    false,
					PlatformType: "centreon",
					CentreonSettings: &v1alpha1.PlatformSpecCentreonSettings{
						URL:                   "http://localhost",
						SelfSignedCertificate: true,
						Secret:                key.Name,
						Endpoint: &v1alpha1.CentreonSpecEndpoint{
							Template:     "template",
							DefaultHost:  "localhost",
							NameTemplate: "ping",
							Macros: map[string]string{
								"mac1": "value1",
								"mac2": "value2",
							},
							Arguments:       []string{"arg1", "arg2"},
							ActivateService: true,
							ServiceGroups:   []string{"sg1"},
							Categories:      []string{"cat1"},
						},
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

			assert.NotEmpty(t, p.Status.SecretHash)
			assert.NotEmpty(t, platforms["test"])
			assert.Equal(t, "test", platforms["test"].platform.Spec.Name)
			assert.NotNil(t, platforms["test"].client)
			assert.NotEmpty(t, platforms["test"].hash)

			data["platform"] = platforms["test"]
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

			assert.Equal(t, "http://localhost2", platforms["test"].platform.Spec.CentreonSettings.URL)
			assert.NotEqual(t, platforms["test"].hash, platform.hash)
			assert.NotEqual(t, platforms["test"].client, platform.client)
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
			p := o.(*v1alpha1.Platform)

			data["version"] = p.ResourceVersion
			data["hash"] = p.Status.SecretHash

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

			d, err = helper.Get(data, "version")
			if err != nil {
				t.Fatal(err)
			}
			version := d.(string)

			d, err = helper.Get(data, "hash")
			if err != nil {
				t.Fatal(err)
			}
			hash := d.(string)

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

			assert.Equal(t, "http://localhost2", platforms["test"].platform.Spec.CentreonSettings.URL)
			assert.NotEqual(t, p.Status.SecretHash, hash)
			assert.NotEqual(t, platforms["test"].client, platform.client)
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
			assert.Nil(t, platforms["test"])

			return nil
		},
	}
}
