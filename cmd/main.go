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
	"crypto/tls"
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/utils/ptr"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	version  = "develop"
	commit   = ""
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(monitorapi.AddToScheme(scheme))

	utilruntime.Must(routev1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var secureMetrics bool
	var probeAddr string
	var tlsOpts []func(*tls.Config)
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	opts := zap.Options{
		Development: true,
		Level:       helper.GetZapLogLevelFromEnv(),
		Encoder:     helper.GetZapFormatterFromDev(),
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	log := logrus.New()
	log.SetLevel(helper.GetLogrusLogLevelFromEnv())
	log.SetFormatter(helper.GetLogrusFormatterFromEnv())

	// Log panics error and exit
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Panic: %v", r)
			os.Exit(1)
		}
	}()

	var cacheNamespaces map[string]cache.Config
	watchNamespace, err := helper.GetWatchNamespaceFromEnv()
	if err != nil {
		setupLog.Info("WATCH_NAMESPACES env variable not setted, the manager will watch and manage resources in all namespaces")
	} else {
		setupLog.Info("Manager look only resources on namespaces %s", watchNamespace)
		watchNamespaces := helper.StringToSlice(watchNamespace, ",")
		cacheNamespaces = make(map[string]cache.Config)
		for _, namespace := range watchNamespaces {
			cacheNamespaces[namespace] = cache.Config{}
		}
	}

	helper.PrintVersion(ctrl.Log, metricsAddr, probeAddr)
	log.Infof("opensearch-operator version: %s - %s", version, commit)

	cfg := ctrl.GetConfigOrDie()
	timeout, err := helper.GetKubeClientTimeoutFromEnv()
	if err != nil {
		setupLog.Error(err, "KUBE_CLIENT_TIMEOUT must be a valid duration: %s", err.Error())
		os.Exit(1)
	}
	cfg.Timeout = timeout

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := server.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:  scheme,
		Metrics: metricsServerOptions,
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: 9443,
		}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "bf9480d8.k8s.harmonie-mutuelle.fr",
		Cache: cache.Options{
			DefaultNamespaces: cacheNamespaces,
		},
		Controller: config.Controller{
			MaxConcurrentReconciles: 1,
			RecoverPanic:            ptr.To(true),
		},

		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Set indexers
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &monitorapi.Platform{}, "spec.centreonSettings.secret", func(o client.Object) []string {
		p := o.(*monitorapi.Platform)
		return []string{p.Spec.CentreonSettings.Secret}
	}); err != nil {
		setupLog.Error(err, "unable to create indexers", "indexers", "Platform")
		os.Exit(1)
	}

	/*

		// Get platforms
		// Not block if errors, maybee not yet platform available
		platforms, err := controllers.ComputedPlatformList(context.Background(), dynamic.NewForConfigOrDie(cfg), kubernetes.NewForConfigOrDie(cfg), log.WithFields(logrus.Fields{
			"type": "CentreonHandler",
		}))
		if err != nil {
			log.Errorf("Error when get platforms, we start controller with empty platform list: %s", err.Error())
			platforms = map[string]*controllers.ComputedPlatform{}
		}
		// Set platform controllers
		platformController := controllers.NewPlatformReconciler(mgr.GetClient(), mgr.GetScheme())
		platformController.SetLogger(log.WithFields(logrus.Fields{
			"type": "PlatformController",
		}))
		platformController.SetRecorder(mgr.GetEventRecorderFor("platform-controller"))
		platformController.SetReconsiler(platformController)
		platformController.SetPlatforms(platforms)
		if err = platformController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Platform")
			os.Exit(1)
		}

		// Template controller sub system
		templateController := controllers.TemplateController{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}
		templateController.SetLogger(log.WithFields(logrus.Fields{
			"type": "TemplateController",
		}))

		// Set CentreonService controller
		centreonServiceController := controllers.NewCentreonServiceReconciler(mgr.GetClient(), mgr.GetScheme())
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

		// Set CentreonServiceGroup controller
		centreonServiceGroupController := controllers.NewCentreonServiceGroupReconciler(mgr.GetClient(), mgr.GetScheme())
		centreonServiceGroupController.SetLogger(log.WithFields(logrus.Fields{
			"type": "CentreonServiceGroupController",
		}))
		centreonServiceGroupController.SetRecorder(mgr.GetEventRecorderFor("centreonservicegroup-controller"))
		centreonServiceGroupController.SetReconsiler(centreonServiceGroupController)
		centreonServiceGroupController.SetPlatforms(platforms)
		if err = centreonServiceGroupController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "CentreonServiceGroup")
			os.Exit(1)
		}

		// Set Ingress controller
		ingressController := controllers.NewIngressReconciler(mgr.GetClient(), mgr.GetScheme(), templateController)
		ingressController.Reconciler.SetLogger(log.WithFields(logrus.Fields{
			"type": "IngressController",
		}))
		ingressController.SetRecorder(mgr.GetEventRecorderFor("ingress-controller"))
		ingressController.SetReconsiler(ingressController)
		ingressController.SetPlatforms(platforms)
		if err = ingressController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Ingress")
			os.Exit(1)
		}

		// Set route controller
		isRouteCRD, err := controllers.IsRouteCRD(cfg)
		if err != nil {
			setupLog.Error(err, "unable to check API groups")
			os.Exit(1)
		}
		if isRouteCRD {
			routeController := controllers.NewRouteReconciler(mgr.GetClient(), mgr.GetScheme(), templateController)
			routeController.Reconciler.SetLogger(log.WithFields(logrus.Fields{
				"type": "RouteController",
			}))
			routeController.SetRecorder(mgr.GetEventRecorderFor("route-controller"))
			routeController.SetReconsiler(routeController)
			routeController.SetPlatforms(platforms)
			if err = routeController.SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", "Route")
				os.Exit(1)
			}
		}

		// Set namespace
		namespaceController := controllers.NewNamespaceReconciler(mgr.GetClient(), mgr.GetScheme(), templateController)
		namespaceController.Reconciler.SetLogger(log.WithFields(logrus.Fields{
			"type": "NamespaceController",
		}))
		namespaceController.SetRecorder(mgr.GetEventRecorderFor("namespace-controller"))
		namespaceController.SetReconsiler(namespaceController)
		namespaceController.SetPlatforms(platforms)
		if err = namespaceController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Namespace")
			os.Exit(1)
		}

		// Set node
		nodeController := controllers.NewNodeReconciler(mgr.GetClient(), mgr.GetScheme(), templateController)
		nodeController.Reconciler.SetLogger(log.WithFields(logrus.Fields{
			"type": "NodeController",
		}))
		nodeController.SetRecorder(mgr.GetEventRecorderFor("node-controller"))
		nodeController.SetReconsiler(nodeController)
		nodeController.SetPlatforms(platforms)
		if err = nodeController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Node")
			os.Exit(1)
		}

		// Set certificate
		certificateController := controllers.NewCertificateReconciler(mgr.GetClient(), mgr.GetScheme(), templateController)
		certificateController.Reconciler.SetLogger(log.WithFields(logrus.Fields{
			"type": "CertificateController",
		}))
		certificateController.SetRecorder(mgr.GetEventRecorderFor("certificate-controller"))
		certificateController.SetReconsiler(certificateController)
		certificateController.SetPlatforms(platforms)
		if err = certificateController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Certificate")
			os.Exit(1)
		}
	*/

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
