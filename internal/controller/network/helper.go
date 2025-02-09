package network

import (
	"context"

	"emperror.dev/errors"
	"github.com/disaster37/k8s-objectmatcher/patch"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
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

//+kubebuilder:rbac:groups="networking.k8s.io",resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete

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
			Labels: map[string]string{
				"app.kubernetes.io/name": "monitoring-operator",
			},
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
		// Create
		if k8serrors.IsNotFound(err) {
			// Set diff 3-way annotations
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(networkPolicy); err != nil {
				return errors.Wrap(err, "Error when set annotation for 3-way diff on NetworkPolicy for webhook")
			}
			if err = c.Create(context.Background(), expectedNetworkPolicy); err != nil {
				return errors.Wrap(err, "Error when create NetworkPolicy for webhook")
			}

			logger.Info("Successfully create networkPolicy for webhook")
			return nil
		}
		return errors.Wrap(err, "Error when get NetworkPolicy for webhook")
	}

	// Diff
	controller.MustInjectTypeMeta(networkPolicy, expectedNetworkPolicy)
	patchResult, err := patch.DefaultPatchMaker.Calculate(networkPolicy, expectedNetworkPolicy)
	if err != nil {
		return errors.Wrap(err, "Error when diffing NetworkPolicy for webhook")
	}

	// Update
	if !patchResult.IsEmpty() {
		networkPolicy = patchResult.Patched.(*networkv1.NetworkPolicy)
		if err = c.Update(context.Background(), networkPolicy); err != nil {
			return errors.Wrap(err, "Error when update NetworkPolicy for webhook")
		}
		logger.Info("Successfully update networkPolicy for webhook")
	}

	return nil
}
