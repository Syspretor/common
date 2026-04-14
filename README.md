# common

A shared library for Fluid workload controllers, providing AdvancedStatefulSet and utility packages for writing Kubernetes controllers.

## Feature

- **AdvancedStatefulSet**: Enhanced StatefulSet with advanced features like in-place update, lifecycle hooks, and more flexible update strategies.

## Usage

### Add Controller

To add the AdvancedStatefulSet controller to your manager:

```go
package main

import (
    "sigs.k8s.io/controller-runtime/pkg/manager"
    
    "github.com/fluid-cloudnative/common/pkg/workload/advancedstatefulset"
)

func main() {
    mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
        // ...
    })
    if err != nil {
        panic(err)
    }

    // Add AdvancedStatefulSet controller
    if err := advancedstatefulset.Add(mgr); err != nil {
        panic(err)
    }

    // Start the manager
    if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
        panic(err)
    }
}
```

### Create and Manage Resources

To create and manage AdvancedStatefulSet resources programmatically:

```go
import (
    "context"
    
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    workloadv1alpha1 "github.com/fluid-cloudnative/common/api/workload/v1alpha1"
)

func createAdvancedStatefulSet(c client.Client) error {
    asts := &workloadv1alpha1.AdvancedStatefulSet{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "example",
            Namespace: "default",
        },
        Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
            StatefulSetSpec: appsv1.StatefulSetSpec{
                Replicas: pointer.Int32(3),
                Selector: &metav1.LabelSelector{
                    MatchLabels: map[string]string{"app": "example"},
                },
                Template: corev1.PodTemplateSpec{
                    // ... pod template
                },
            },
            // Advanced features
            UpdateStrategy: workloadv1alpha1.AdvancedStatefulSetUpdateStrategy{
                Type: workloadv1alpha1.RollingUpdateAdvancedStatefulSetStrategyType,
                RollingUpdate: &workloadv1alpha1.RollingUpdateStatefulSetStrategy{
                    PodUpdatePolicy: workloadv1alpha1.InPlaceIfPossiblePodUpdateStrategyType,
                },
            },
        },
    }
    
    return c.Create(context.TODO(), asts)
}
```

## Permissions

The controller requires the following RBAC permissions:

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: advancedstatefulset-controller
rules:
  # Core resources
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - get
      - list
      - watch
      - create
      - update
      - patch
      - delete
  - apiGroups:
      - ""
    resources:
      - pods/status
    verbs:
      - get
      - update
      - patch
  - apiGroups:
      - ""
    resources:
      - events
    verbs:
      - get
      - list
      - watch
      - create
      - update
      - patch
      - delete
  - apiGroups:
      - ""
    resources:
      - persistentvolumeclaims
    verbs:
      - get
      - list
      - watch
      - create
      - update
      - patch
      - delete
  # Apps resources
  - apiGroups:
      - apps
    resources:
      - controllerrevisions
    verbs:
      - get
      - list
      - watch
      - create
      - update
      - patch
      - delete
  # Custom resources
  - apiGroups:
      - workload.fluid.io
    resources:
      - advancedstatefulsets
    verbs:
      - get
      - list
      - watch
      - create
      - update
      - patch
      - delete
  - apiGroups:
      - workload.fluid.io
    resources:
      - advancedstatefulsets/status
    verbs:
      - get
      - update
      - patch
  - apiGroups:
      - workload.fluid.io
    resources:
      - advancedstatefulsets/finalizers
    verbs:
      - update
```

## Healthz and Readyz

The controller integrates with controller-runtime's health and readiness checks automatically. When using `advancedstatefulset.Add(mgr)`, the controller will:

- Report **ready** once the controller is initialized and caches are synced
- Report **healthy** as long as the reconcile loop is functioning

To add custom health checks:

```go
import (
    "net/http"
    "sigs.k8s.io/controller-runtime/pkg/healthz"
)

// Add liveness probe
_ = mgr.AddHealthzCheck("custom-check", func(req *http.Request) error {
    // Custom health check logic
    return nil
})

// Add readiness probe
_ = mgr.AddReadyzCheck("custom-ready", func(req *http.Request) error {
    // Custom readiness check logic
    return nil
})
```

Configure probes in your deployment:

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8081
  initialDelaySeconds: 15
  periodSeconds: 20
readinessProbe:
  httpGet:
    path: /readyz
    port: 8081
  initialDelaySeconds: 5
  periodSeconds: 10
```

## Installation

### As a Dependency

```bash
go get github.com/fluid-cloudnative/common/pkg/workload
```

### Import

```go
import (
    // API types
    workloadv1alpha1 "github.com/fluid-cloudnative/common/api/workload/v1alpha1"
    
    // Controller
    "github.com/fluid-cloudnative/common/pkg/workload/advancedstatefulset"
    
    // Utilities
    "github.com/fluid-cloudnative/common/pkg/workload/utils/inplaceupdate"
    "github.com/fluid-cloudnative/common/pkg/workload/utils/lifecycle"
)
```

## License

Apache License 2.0
