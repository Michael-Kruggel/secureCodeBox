// SPDX-FileCopyrightText: the secureCodeBox authors
//
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"context"
	"fmt"

	v1 "github.com/secureCodeBox/secureCodeBox/operator/apis/execution/v1"
	"github.com/secureCodeBox/secureCodeBox/operator/utils"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewTriggerCommand() *cobra.Command {
	triggerCmd := &cobra.Command{
		Use:          "trigger [filename]",
		Short:        "Trigger a new scheduled scan",
		Long:         `Trigger a new scan custom resource in the current namespace`,
		Example:      ``,
		SilenceUsage: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("you must specify a scheduled scan name")
			}
			scheduledScanName := args[0]
			kubeclient, namespace, err := clientProvider.GetClient(kubeconfigArgs)
			if err != nil {
				return fmt.Errorf("error initializing kubernetes client, your kubeconfig is likely malformed or invalid")
			}

			if namespaceFlag, err := cmd.Flags().GetString("namespace"); err == nil && namespaceFlag != "" {
				namespace = namespaceFlag
			}

			var scan v1.ScheduledScan

			err = kubeclient.Get(context.TODO(), types.NamespacedName{Name: scheduledScanName, Namespace: namespace}, &scan)
			if err != nil {
				if apierrors.IsNotFound(err) {
					return fmt.Errorf("could not find ScheduledScan '%s' in namespace '%s'", scheduledScanName, namespace)
				} else {
					panic(err)
				}
			}

			utils.RetriggerScheduledScan(context.TODO(), kubeclient.Status(), scan)
			fmt.Printf("triggered new Scan for ScheduledScan '%s'\n", scheduledScanName)

			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			kubeclient, namespace, err := clientProvider.GetClient(kubeconfigArgs)
			if err != nil {
				fmt.Printf("Error initializing kubernetes client: %v\n", err)
				return nil, cobra.ShellCompDirectiveError
			}

			if namespaceFlag, err := cmd.Flags().GetString("namespace"); err == nil && namespaceFlag != "" {
				namespace = namespaceFlag
			}

			var scans v1.ScheduledScanList

			err = kubeclient.List(context.TODO(), &scans, client.InNamespace(namespace))
			if err != nil {
				fmt.Printf("Error listing ScheduledScans: %v\n", err)
				return nil, cobra.ShellCompDirectiveError
			}

			scanNames := make([]string, len(scans.Items))
			for i, scan := range scans.Items {
				scanNames[i] = scan.Name
			}

			return scanNames, cobra.ShellCompDirectiveDefault
		},
	}

	return triggerCmd
}
