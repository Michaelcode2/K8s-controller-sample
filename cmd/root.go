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
	Use:   "k8s-controller-tutorial",
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

Examples:
  # Monitor deployments in default namespace
  k8s-controller-tutorial controller

  # Watch for changes in kube-system namespace
  k8s-controller-tutorial controller -n kube-system -w

  # Get help for controller command
  k8s-controller-tutorial controller --help`,
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
