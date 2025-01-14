package network

import (
	"context"

	"emperror.dev/errors"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateNetworkPolicyForWebhook permit to create / update NetworkPolicy for webhook
func CreateNetworkPolicyForWebhook(c client.Client, logger *logrus.Entry) error {
	namespace, err := helpers.GetOperatorNamespace()
	if err != nil {
		return err
	}

	networkPolicy := &networkv1.NetworkPolicy{}
	expectedNetworkPolicy := &networkv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "allow-webhook-access-from-any",
		},
		Spec: networkv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"control-plane": "monitoring-operator",
				},
			},
			Ingress: []networkv1.NetworkPolicyIngressRule{
				{
					From: []networkv1.NetworkPolicyPeer{},
					Ports: []networkv1.NetworkPolicyPort{
						{
							Protocol: ptr.To(v1.ProtocolTCP),
							Port:     ptr.To(intstr.FromInt(9443)),
						},
					},
				},
			},
		},
	}

	if err = c.Get(context.Background(), types.NamespacedName{Namespace: expectedNetworkPolicy.GetNamespace(), Name: expectedNetworkPolicy.GetName()}, networkPolicy); err != nil {
		if k8serrors.IsNotFound(err) {
			if err = c.Create(context.Background(), expectedNetworkPolicy); err != nil {
				return errors.Wrap(err, "Error when create NetworkPolicy for webhook")
			}

			logger.Info("Successfully create networkPolicy for webhook")
			return nil
		}
		return errors.Wrap(err, "Error when get NetworkPolicy for webhook")
	}

	if !cmp.Equal(networkPolicy.Spec, expectedNetworkPolicy.Spec) {
		networkPolicy.Spec = expectedNetworkPolicy.Spec
		if c.Update(context.Background(), networkPolicy); err != nil {
			return errors.Wrap(err, "Error when update NetworkPolicy for webhook")
		}
		logger.Info("Successfully update networkPolicy for webhook")
	}

	return nil
}
