/*
Copyright © 2025 Michael Vaynagiy (Michaelcode2)
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8s-controller-sample",
	Short: "A Kubernetes controller for monitoring deployments and events",
	Long: `A comprehensive Kubernetes controller built with Go and Cobra that provides
real-time monitoring and management of Kubernetes deployments.

Features:
  • Monitor deployment status and health
  • Watch for real-time deployment changes
  • Track Kubernetes events and notifications
  • Structured logging with environment support
  • Namespace-specific monitoring
  • Production-ready with JSON logging
  • HTTP API server with FastHTTP
  • Controller-runtime with leader election
  • Metrics and health endpoints

Examples:
  # Monitor deployments in default namespace
  k8s-controller-sample controller

  # Watch for changes in kube-system namespace
  k8s-controller-sample controller -n kube-system -w

  # Start HTTP server on port 8080
  k8s-controller-sample server -p 8080

  # Start controller-runtime manager with leader election
  k8s-controller-sample manager --leader-elect

  # Start HTTP server on specific host and port
  k8s-controller-sample server -H 127.0.0.1 -p 9090

  # Get help for controller command
  k8s-controller-sample controller --help

  # Get help for server command
  k8s-controller-sample server --help

  # Get help for manager command
  k8s-controller-sample manager --help`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.k8s-controller-tutorial.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
