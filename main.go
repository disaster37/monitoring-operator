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

package main

import (
	"flag"
	"os"
	"sync/atomic"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/disaster37/go-centreon-rest/v21"
	monitorv1alpha1 "github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/controllers"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	//+kubebuilder:scaffold:imports
)

const (
	MONITORING_CENTREON = "centreon"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	version  = "develop"
	commit   = ""
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(monitorv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
		Level:       getZapLogLevel(),
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	log := logrus.New()
	log.SetLevel(getLogrusLogLevel())

	watchNamespace, err := getWatchNamespace()
	var namespace string
	var multiNamespacesCached cache.NewCacheFunc

	if err != nil {
		setupLog.Info("WATCH_NAMESPACES env variable not setted, the manager will watch and manage resources in all namespaces")
	} else {
		setupLog.Info("Manager look only resources on namespaces %s", watchNamespace)
		watchNamespaces := helpers.StringToSlice(watchNamespace, ",")
		if len(watchNamespaces) == 1 {
			namespace = watchNamespace
		} else {
			multiNamespacesCached = cache.MultiNamespacedCacheBuilder(watchNamespaces)
		}

	}

	printVersion(ctrl.Log, metricsAddr, probeAddr)
	log.Infof("monitoring-operator version: %s - %s", version, commit)

	cfg := ctrl.GetConfigOrDie()
	timeout, err := getKubeClientTimeout()
	if err != nil {
		setupLog.Error(err, "KUBE_CLIENT_TIMEOUT must be a valid duration: %s", err.Error())
		os.Exit(1)
	}
	cfg.Timeout = timeout

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "351cbbed.k8s.webcenter.fr",
		Namespace:              namespace,
		NewCache:               multiNamespacesCached,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Get monitoring type to enable the right controllers
	monitoringType := os.Getenv("MONITORING_PLATEFORM")
	switch monitoringType {
	case MONITORING_CENTREON:
		// Init Centreon API client
		cfg, err := helpers.GetCentreonConfig()
		if err != nil {
			setupLog.Error(err, "unable to get Centreon config")
			os.Exit(1)
		}
		if log.GetLevel() == logrus.DebugLevel {
			cfg.Debug = true
		}
		centreonClient, err := centreon.NewClient(cfg)
		if err != nil {
			setupLog.Error(err, "unable to get Centreon client")
			os.Exit(1)
		}
		centreonHandler := centreonhandler.NewCentreonHandler(centreonClient, logrus.NewEntry(log))

		// Init CentreonConfig
		var a atomic.Value

		// Set controllers for Centreon resources
		if err = (&controllers.CentreonServiceReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			Service:        controllers.NewCentreonService(centreonHandler),
			CentreonConfig: &a,
			Log: log.WithFields(logrus.Fields{
				"type": "CentreonServiceController",
			}),
			Recorder: mgr.GetEventRecorderFor("centreonservice-controller"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "CentreonService")
			os.Exit(1)
		}

		if err = (&controllers.IngressCentreonReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			CentreonConfig: &a,
			Recorder:       mgr.GetEventRecorderFor("ingresscentreon-controller"),
			Log: log.WithFields(logrus.Fields{
				"type": "IngressCentreonControllers",
			}),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "IngressCentreon")
			os.Exit(1)
		}

		if err = (&controllers.CentreonReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			CentreonConfig: &a,
			Recorder:       mgr.GetEventRecorderFor("centreon-controller"),
			Log: log.WithFields(logrus.Fields{
				"type": "CentreonControllers",
			}),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Centreon")
			os.Exit(1)
		}

		break
	default:
		setupLog.Error(errors.Errorf("MONITORING_PLATEFORM not supported. You need to set %s", MONITORING_CENTREON), "Monitoring plateform not supported")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
