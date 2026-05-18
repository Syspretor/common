/*
Copyright 2026 The Fluid Authors.

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
	"fmt"
	"os"
	"time"

	"github.com/fluid-cloudnative/advanced-statefulset/api/workload/v1alpha1"
	"github.com/fluid-cloudnative/advanced-statefulset/pkg/workload/advancedstatefulset"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	metricsAddr          = flag.String("metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	probeAddr            = flag.String("health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	enableLeaderElection = flag.Bool("leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	createExample = flag.Bool("create-example", false, "Create an example AdvancedStatefulSet on startup")
	namespace     = flag.String("namespace", "default", "Namespace for the example AdvancedStatefulSet")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	ctx := signals.SetupSignalHandler()

	// Get config
	cfg, err := config.GetConfig()
	if err != nil {
		// Try to load from kubeconfig for local development
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		}
		cfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			&clientcmd.ConfigOverrides{},
		).ClientConfig()
		if err != nil {
			klog.ErrorS(err, "Unable to get kubeconfig")
			os.Exit(1)
		}
	}

	// Setup scheme
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)

	// Setup manager
	mgr, err := manager.New(cfg, manager.Options{
		Scheme:           scheme,
		LeaderElection:   *enableLeaderElection,
		LeaderElectionID: "advancedstatefulset-controller.fluid.io",
	})
	if err != nil {
		klog.ErrorS(err, "Unable to create manager")
		os.Exit(1)
	}

	// Add AdvancedStatefulSet controller
	klog.InfoS("Setting up AdvancedStatefulSet controller")
	if err := advancedstatefulset.Add(mgr); err != nil {
		klog.ErrorS(err, "Unable to add AdvancedStatefulSet controller")
		os.Exit(1)
	}

	// Create example AdvancedStatefulSet if requested
	if *createExample {
		go func() {
			// Wait for manager to start
			time.Sleep(5 * time.Second)
			if err := createExampleAdvancedStatefulSet(mgr.GetClient(), *namespace); err != nil {
				klog.ErrorS(err, "Failed to create example AdvancedStatefulSet")
			}
		}()
	}

	// Start manager
	klog.InfoS("Starting manager", "metricsAddr", *metricsAddr, "probeAddr", *probeAddr)
	if err := mgr.Start(ctx); err != nil {
		klog.ErrorS(err, "Problem running manager")
		os.Exit(1)
	}
}

// createExampleAdvancedStatefulSet creates an example AdvancedStatefulSet
func createExampleAdvancedStatefulSet(c client.Client, ns string) error {
	ctx := context.Background()

	// Check if namespace exists, create if not
	namespaceObj := &corev1.Namespace{}
	namespaceObj.Name = ns
	if err := c.Create(ctx, namespaceObj); err != nil {
		klog.V(4).InfoS("Namespace may already exist", "namespace", ns, "error", err)
	}

	replicas := int32(3)
	example := &v1alpha1.AdvancedStatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-asts",
			Namespace: ns,
			Labels: map[string]string{
				"app": "example",
			},
		},
		Spec: v1alpha1.AdvancedStatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "example",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "example",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:alpine",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Name:          "http",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
			},
			ServiceName: "example-asts-headless",
			UpdateStrategy: v1alpha1.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &v1alpha1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: v1alpha1.InPlaceIfPossiblePodUpdateStrategyType,
				},
			},
		},
	}

	// Try to create the AdvancedStatefulSet
	if err := c.Create(ctx, example); err != nil {
		// If it already exists, update it
		existing := &v1alpha1.AdvancedStatefulSet{}
		existing.Name = example.Name
		existing.Namespace = example.Namespace
		if err := c.Get(ctx, client.ObjectKeyFromObject(existing), existing); err != nil {
			return fmt.Errorf("failed to get existing AdvancedStatefulSet: %w", err)
		}
		example.ResourceVersion = existing.ResourceVersion
		if err := c.Update(ctx, example); err != nil {
			return fmt.Errorf("failed to update AdvancedStatefulSet: %w", err)
		}
		klog.InfoS("Updated example AdvancedStatefulSet", "name", example.Name, "namespace", example.Namespace)
	} else {
		klog.InfoS("Created example AdvancedStatefulSet", "name", example.Name, "namespace", example.Namespace)
	}

	// List and display AdvancedStatefulSets
	astsList := &v1alpha1.AdvancedStatefulSetList{}
	if err := c.List(ctx, astsList, client.InNamespace(ns)); err != nil {
		klog.ErrorS(err, "Failed to list AdvancedStatefulSets")
	} else {
		klog.InfoS("AdvancedStatefulSets in namespace", "namespace", ns, "count", len(astsList.Items))
		for _, item := range astsList.Items {
			selector, _ := metav1.LabelSelectorAsSelector(item.Spec.Selector)
			klog.InfoS("Found AdvancedStatefulSet",
				"name", item.Name,
				"replicas", item.Spec.Replicas,
				"selector", selector.String(),
			)
		}
	}

	return nil
}

// Helper function for pointer
func int32Ptr(i int32) *int32 {
	return &i
}

// Helper function to get selector string
func getSelectorString(selector *metav1.LabelSelector) string {
	s, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return "<invalid>"
	}
	return s.String()
}

// Helper function for labels
func mustParseSelector(s string) labels.Selector {
	selector, err := labels.Parse(s)
	if err != nil {
		panic(err)
	}
	return selector
}
