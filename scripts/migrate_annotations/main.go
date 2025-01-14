package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"emperror.dev/errors"
	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/google/go-cmp/cmp"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	version string
	commit  string
)

func run(args []string) error {
	// Logger setting
	formatter := new(prefixed.TextFormatter)
	formatter.FullTimestamp = true
	formatter.ForceFormatting = true
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)

	// Get home directory
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Warnf("Can't get home directory: %s", err.Error())
		homePath = "/root"
	}

	// CLI settings
	app := cli.NewApp()
	app.Usage = "Manage Opensearch on cli interface"
	app.Version = fmt.Sprintf("%s-%s", version, commit)
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "Load configuration from `FILE`",
		},
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "kubeconfig",
			Usage:   "The kube config file",
			EnvVars: []string{"KUBECONFIG"},
			Value:   fmt.Sprintf("%s/.kube/config", homePath),
		}),
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Display debug output",
		},
		altsrc.NewInt64Flag(&cli.Int64Flag{
			Name:  "timeout",
			Usage: "The timeout in second",
			Value: 0,
		}),
		&cli.BoolFlag{
			Name:  "no-color",
			Usage: "No print color",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:  "migrate",
			Usage: "Migrate annotations",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "old",
					Usage:    "The old annotation to migrate",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "new",
					Usage:    "The new annotation to replace by",
					Required: true,
				},
			},
			Action: migrateAnnotations,
		},
	}

	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		if !c.Bool("no-color") {
			formatter := new(prefixed.TextFormatter)
			formatter.FullTimestamp = true
			formatter.ForceFormatting = true
			log.SetFormatter(formatter)
		}

		if c.String("config") != "" {
			before := altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewYamlSourceFromFlagFunc("config"))
			return before(c)
		}
		return nil
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	return app.Run(args)
}

func main() {
	err := run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func migrateAnnotations(c *cli.Context) error {
	var (
		old *types.NamespacedName
		new *types.NamespacedName
	)

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(centreoncrd.AddToScheme(scheme))
	utilruntime.Must(routev1.AddToScheme(scheme))

	logger := log.NewEntry(log.New())

	cfg, err := clientcmd.BuildConfigFromFlags("", c.String("kubeconfig"))
	if err != nil {
		return errors.Wrap(err, "Error when load kubeconfig")
	}
	client, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	// Pasre params
	if err := json.Unmarshal([]byte(c.String("old")), old); err != nil {
		return errors.Wrap(err, "Error when unmarhall old typedNamespace")
	}
	if err := json.Unmarshal([]byte(c.String("new")), new); err != nil {
		return errors.Wrap(err, "Error when unmarhall new typedNamespace")
	}

	// Namespaces
	logger.Info("Start migrate resources of kind Namespace")
	namespaces := &corev1.NamespaceList{}
	if err := client.List(c.Context, namespaces); err != nil {
		return errors.Wrap(err, "Error when list all namespaces")
	}
	if err := doMigrateAnnotations(c.Context, *old, *new, namespaces, client, logger); err != nil {
		return err
	}

	// Certificates
	logger.Info("Start migrate resources of kind Secret (certificate)")
	secrets := &corev1.SecretList{}
	if err := client.List(c.Context, secrets); err != nil {
		return errors.Wrap(err, "Error when list all secrets")
	}
	if err := doMigrateAnnotations(c.Context, *old, *new, secrets, client, logger); err != nil {
		return err
	}

	// Nodes
	logger.Info("Start migrate resources of kind Node")
	nodes := &corev1.NodeList{}
	if err := client.List(c.Context, nodes); err != nil {
		return errors.Wrap(err, "Error when list all nodes")
	}
	if err := doMigrateAnnotations(c.Context, *old, *new, nodes, client, logger); err != nil {
		return err
	}

	// Ingress
	logger.Info("Start migrate resources of kind Ingress")
	ingresses := &networkv1.IngressList{}
	if err := client.List(c.Context, ingresses); err != nil {
		return errors.Wrap(err, "Error when list all ingresses")
	}
	if err := doMigrateAnnotations(c.Context, *old, *new, ingresses, client, logger); err != nil {
		return err
	}

	// Routes
	logger.Info("Start migrate resources of kind Ingress")
	routes := &routev1.RouteList{}
	if err := client.List(c.Context, routes); err != nil {
		return errors.Wrap(err, "Error when list all routes")
	}
	if err := doMigrateAnnotations(c.Context, *old, *new, routes, client, logger); err != nil {
		return err
	}

	return nil
}

func doMigrateAnnotations(ctx context.Context, old types.NamespacedName, new types.NamespacedName, list client.ObjectList, client client.Client, logger *logrus.Entry) error {
	listNamespacedName := make([]types.NamespacedName, 0)

	for _, o := range helpers.GetItems(list) {
		logger = logger.WithFields(log.Fields{
			"kind":      o.GetObjectKind().GroupVersionKind().Kind,
			"namespace": o.GetNamespace(),
			"name":      o.GetName(),
		})
		log.Debug("Start process resource")

		targetTemplates := o.GetAnnotations()[fmt.Sprintf("%s/templates", centreoncrd.MonitoringAnnotationKey)]
		if targetTemplates != "" {
			if err := json.Unmarshal([]byte(targetTemplates), &listNamespacedName); err != nil {
				return errors.Wrap(err, "Error when unmarshall the list of template")
			}
		}

		needUpdate := false
		for i, namespacedName := range listNamespacedName {
			logger.Debugf("Found resource with annotations %s", namespacedName.String())
			if cmp.Equal(old, namespacedName) {
				logger.Debug("Found annotation to migrate")
				listNamespacedName[i] = new
				needUpdate = true
			}
		}
		if needUpdate {

			b, err := json.Marshal(listNamespacedName)
			if err != nil {
				return errors.Wrap(err, "Error when marshall annotations")
			}

			annotations := o.GetAnnotations()
			annotations[fmt.Sprintf("%s/templates", centreoncrd.MonitoringAnnotationKey)] = string(b)
			o.SetAnnotations(annotations)

			if err := client.Update(ctx, o); err != nil {
				return errors.Wrap(err, "Error when update object")
			}
		}
	}

	return nil
}
