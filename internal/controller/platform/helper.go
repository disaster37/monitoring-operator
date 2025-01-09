package platform

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"emperror.dev/errors"
	"github.com/disaster37/go-centreon-rest/v21"
	"github.com/disaster37/go-centreon-rest/v21/models"
	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetClient premit to get client to connect on monitoring platform
func GetClient(platformRef string, platforms map[string]*ComputedPlatform) (meta any, platform *monitorapi.Platform, err error) {
	if platformRef == "" {
		if p, ok := platforms["default"]; ok {
			return p.Client, p.Platform, nil
		}

		return nil, nil, errors.New("No default platform")
	}

	if p, ok := platforms[platformRef]; ok {
		return p.Client, p.Platform, nil
	}

	return nil, nil, errors.Errorf("Platform %s not found", platformRef)
}

// ComputedPlatformList permit to get the list of coomputed platform object
// It usefull to init controller with client to access on external monitoring resources
func ComputedPlatformList(ctx context.Context, c client.Client, logger *logrus.Entry) (platforms map[string]*ComputedPlatform, err error) {
	platforms = map[string]*ComputedPlatform{}
	platformList := &monitorapi.PlatformList{}
	ns, err := helpers.GetOperatorNamespace()
	if err != nil {
		return nil, err
	}

	// Get list of current platform
	if err = c.List(ctx, platformList, &client.ListOptions{Namespace: ns}); err != nil {
		return nil, err
	}

	// Create computed platform
	for _, p := range platformList.Items {
		logger.Debugf("Start to Compute platform %s", p.Name)

		switch p.Spec.PlatformType {
		case "centreon":
			// Get centreon secret
			s := &corev1.Secret{}
			k := types.NamespacedName{
				Namespace: p.Namespace,
				Name:      p.Spec.CentreonSettings.Secret,
			}
			if err = c.Get(ctx, k, s); err != nil {
				if k8serrors.IsNotFound(err) {
					logger.Warnf("Secret %s not yet exist, skip it", p.Spec.CentreonSettings.Secret)
					continue
				}
			}
			cp, err := getComputedCentreonPlatform(&p, s, logger)
			if err != nil {
				return nil, errors.Wrapf(err, "Error when compute platform %s", p.Name)
			}
			platforms[p.Name] = cp
			if p.Spec.IsDefault {
				platforms["default"] = cp
			}

		default:
			return nil, errors.Errorf("Platform %s of type %s is not supported", p.Name, p.Spec.PlatformType)

		}
	}

	return platforms, nil
}

func getComputedCentreonPlatform(p *monitorapi.Platform, s *corev1.Secret, log *logrus.Entry) (cp *ComputedPlatform, err error) {
	if p == nil {
		return nil, errors.New("Platform can't be null")
	}
	if s == nil {
		return nil, errors.New("Secret can't be null")
	}

	username := string(s.Data["username"])
	password := string(s.Data["password"])
	if username == "" || password == "" {
		return nil, errors.Errorf("You need to set username and password on secret %s", s.Name)
	}

	// Create client
	cfg := &models.Config{
		Address:          p.Spec.CentreonSettings.URL,
		Username:         username,
		Password:         password,
		DisableVerifySSL: p.Spec.CentreonSettings.SelfSignedCertificate,
	}
	if log.Level == logrus.DebugLevel {
		cfg.Debug = true
	}
	client, err := centreon.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Error when create Centreon client")
	}
	shaByte, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	sha := sha256.New()
	if _, err := sha.Write([]byte(shaByte)); err != nil {
		return nil, err
	}

	return &ComputedPlatform{
		Client:   centreonhandler.NewCentreonHandler(client, log),
		Platform: p,
		Hash:     hex.EncodeToString(sha.Sum(nil)),
	}, nil
}
