//go:build !no_server

/*
Copyright 2017 by the contributors.

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
	"fmt"
	"strings"

	"k8s.io/sample-controller/pkg/signals"
	"sigs.k8s.io/aws-iam-authenticator/pkg"
	"sigs.k8s.io/aws-iam-authenticator/pkg/mapper"
	"sigs.k8s.io/aws-iam-authenticator/pkg/metrics"
	"sigs.k8s.io/aws-iam-authenticator/pkg/server"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sigs.k8s.io/aws-iam-authenticator/pkg/endpoints"
)

const (
	// DefaultPort is the default localhost port (chosen randomly).
	DefaultPort = 21362
	// Default Ec2 TPS Variables
	DefaultEC2DescribeInstancesQps   = 15
	DefaultEC2DescribeInstancesBurst = 5
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a webhook validation server suitable that validates tokens using AWS IAM",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		fmt.Printf("Authenticator Version: %q, %q\n", pkg.Version, pkg.CommitID)
		metrics.InitMetrics(prometheus.DefaultRegisterer)

		ctx := signals.SetupSignalHandler()

		cfg, err := getConfig()
		if err != nil {
			logrus.Fatalf("%s", err)
		}

		httpServer := server.New(ctx, cfg)
		httpServer.Run(ctx.Done())
	},
}

func init() {
	serverCmd.Flags().String("partition",
		endpoints.AwsPartitionID,
		fmt.Sprintf("The AWS partition. Must be one of: %v", endpoints.PARTITIONS))
	viper.BindPFlag("server.partition", serverCmd.Flags().Lookup("partition"))

	serverCmd.Flags().String("generate-kubeconfig",
		"/etc/kubernetes/aws-iam-authenticator/kubeconfig.yaml",
		"Output `path` where a generated webhook kubeconfig (for `--authentication-token-webhook-config-file`) will be stored.  When running as a container, this should be a hostPath mount and the API server must be able to access the file.")
	viper.BindPFlag("server.generateKubeconfig", serverCmd.Flags().Lookup("generate-kubeconfig"))

	serverCmd.Flags().Bool("kubeconfig-pregenerated",
		false,
		"Set to `true` when a webhook kubeconfig is pre-generated by running the `init` command, and therefore the `server` shouldn't unnecessarily re-generate a new one.")
	viper.BindPFlag("server.kubeconfigPregenerated", serverCmd.Flags().Lookup("kubeconfig-pregenerated"))

	serverCmd.Flags().String("state-dir",
		"/var/aws-iam-authenticator",
		"State `directory` for generated certificate and private key.  When running as a container, this should be a hostPath mount so that the certificate and key persisted across resarts.")
	viper.BindPFlag("server.stateDir", serverCmd.Flags().Lookup("state-dir"))

	serverCmd.Flags().String("kubeconfig",
		"",
		"kubeconfig file path for using a local kubeconfig to configure the client to talk to the API server for CRD and EKSConfigMap backends.")
	viper.BindPFlag("server.kubeconfig", serverCmd.Flags().Lookup("kubeconfig"))
	serverCmd.Flags().String("master",
		"",
		"master is the URL to the api server, which is merged with the kubeconfig for CRD and EKSConfigMap backends.")
	viper.BindPFlag("server.master", serverCmd.Flags().Lookup("master"))

	serverCmd.Flags().String("address",
		"127.0.0.1",
		"IP Address to bind the aws-iam-authenticator server to listen to. For example: 127.0.0.1 or 0.0.0.0")
	viper.BindPFlag("server.address", serverCmd.Flags().Lookup("address"))

	serverCmd.Flags().StringSlice("backend-mode",
		[]string{mapper.ModeMountedFile},
		fmt.Sprintf("Ordered list of backends to get mappings from. The first one that returns a matching mapping wins. Comma-delimited list of: %s", strings.Join(mapper.BackendModeChoices, ",")))
	viper.BindPFlag("server.backendMode", serverCmd.Flags().Lookup("backend-mode"))

	serverCmd.Flags().Int(
		"port",
		DefaultPort,
		"Port to bind the server to listen to")
	viper.BindPFlag("server.port", serverCmd.Flags().Lookup("port"))

	serverCmd.Flags().Int(
		"ec2-describeInstances-qps",
		DefaultEC2DescribeInstancesQps,
		"AWS EC2 rate limiting with qps")
	viper.BindPFlag("server.ec2DescribeInstancesQps", serverCmd.Flags().Lookup("ec2-describeInstances-qps"))

	serverCmd.Flags().Int(
		"ec2-describeInstances-burst",
		DefaultEC2DescribeInstancesBurst,
		"AWS EC2 rate Limiting with burst")
	viper.BindPFlag("server.ec2DescribeInstancesBurst", serverCmd.Flags().Lookup("ec2-describeInstances-burst"))

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	_ = fs.Parse([]string{})
	flag.CommandLine = fs

	rootCmd.AddCommand(serverCmd)
}
