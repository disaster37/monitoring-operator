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
	"context"
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	monitorv1alpha1 "github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/controllers"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
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

	utilruntime.Must(routev1.AddToScheme(scheme))
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

	// Get platforms
	// Not block if errors, maybee not yet platform available
	platforms, err := controllers.PlatformList(context.Background(), mgr.GetClient(), logrus.NewEntry(log), mgr.GetEventRecorderFor("platform-controller"))
	if err != nil {
		log.Errorf("Error when init platforms: %s", err.Error())
	}

	// Set platform controllers
	platformController := &controllers.PlatformReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	platformController.SetLogger(log.WithFields(logrus.Fields{
		"type": "PlatformController",
	}))
	platformController.SetRecorder(mgr.GetEventRecorderFor("centreonservice-controller"))
	platformController.SetReconsiler(platformController)
	platformController.SetPlatforms(platforms)
	if err = platformController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Platform")
		os.Exit(1)
	}

	// Set secret controller
	secretController := &controllers.SecretReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	secretController.SetLogger(log.WithFields(logrus.Fields{
		"type": "SecretController",
	}))
	secretController.SetRecorder(mgr.GetEventRecorderFor("secret-controller"))
	secretController.SetPlatforms(platforms)
	if err = secretController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Secret")
		os.Exit(1)
	}

	// Set CentreonService controller
	centreonServiceController := &controllers.CentreonServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	centreonServiceController.SetLogger(log.WithFields(logrus.Fields{
		"type": "CentreonServiceController",
	}))
	centreonServiceController.SetRecorder(mgr.GetEventRecorderFor("centreonservice-controller"))
	centreonServiceController.SetReconsiler(centreonServiceController)
	centreonServiceController.SetPlatforms(platforms)
	if err = centreonServiceController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CentreonService")
		os.Exit(1)
	}

	// Set Ingress controller
	ingressController := &controllers.IngressReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	ingressController.SetLogger(log.WithFields(logrus.Fields{
		"type": "IngressCentreonController",
	}))
	ingressController.SetRecorder(mgr.GetEventRecorderFor("ingresscentreon-controller"))
	ingressController.SetReconsiler(ingressController)
	ingressController.SetPlatforms(platforms)
	if err = ingressController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IngressCentreon")
		os.Exit(1)
	}

	// Set route controller
	isRouteCRD, err := controllers.IsRouteCRD(cfg)
	if err != nil {
		setupLog.Error(err, "unable to check API groups")
		os.Exit(1)
	}
	if isRouteCRD {
		routeController := &controllers.RouteReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}
		routeController.SetLogger(log.WithFields(logrus.Fields{
			"type": "RouteCentreonController",
		}))
		routeController.SetRecorder(mgr.GetEventRecorderFor("routecentreon-controller"))
		routeController.SetReconsiler(routeController)
		routeController.SetPlatforms(platforms)
		if err = routeController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "RouteCentreon")
			os.Exit(1)
		}
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
