/*

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

	"github.com/uswitch/nidhogg/pkg/apis"
	"github.com/uswitch/nidhogg/pkg/controller"
	"github.com/uswitch/nidhogg/pkg/nidhogg"
	"github.com/uswitch/nidhogg/pkg/webhook"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	zap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	metricsAddr     string
	configPath      string
	leaderElection  bool
	leaderConfigMap string
	leaderNamespace string
)

func main() {

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&configPath, "config-file", "config.json", "Path to config file")
	flag.BoolVar(&leaderElection, "leader-election", false, "enable leader election")
	flag.StringVar(&leaderConfigMap, "leader-configmap", "", "Name of configmap to use for leader election")
	flag.StringVar(&leaderNamespace, "leader-namespace", "", "Namespace where leader configmap located")
	flag.Parse()
	logf.SetLogger(zap.New(zap.UseDevMode(false)))
	log := logf.Log.WithName("entrypoint")

	handlerConf, err := nidhogg.GetConfig(configPath)
	if err != nil {
		log.Error(err, "unable to get config")
		os.Exit(1)
	}

	log.Info("looking for nodes that match selector", handlerConf.Selector.String())

	// Get a config to talk to the apiserver
	log.Info("setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	log.Info("setting up manager")
	mgr, err := manager.New(cfg, manager.Options{
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		LeaderElection:          leaderElection,
		LeaderElectionID:        leaderConfigMap,
		LeaderElectionNamespace: leaderNamespace,
	})
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	log.Info("setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable add APIs to scheme")
		os.Exit(1)
	}

	// Setup all Controllers
	log.Info("Setting up controller")
	if err := controller.AddToManager(mgr, handlerConf); err != nil {
		log.Error(err, "unable to register controllers to the manager")
		os.Exit(1)
	}

	log.Info("setting up webhooks")
	if err := webhook.AddToManager(mgr); err != nil {
		log.Error(err, "unable to register webhooks to the manager")
		os.Exit(1)
	}

	// Start the Cmd
	log.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "unable to run the manager")
		os.Exit(1)
	}
}
