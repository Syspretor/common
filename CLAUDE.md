# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AdvancedStatefulSet controller and utility packages for Fluid. Provides an enhanced StatefulSet CRD with in-place update, lifecycle hooks, and flexible update strategies. Derived from OpenKruise. The CRD group is `workload.fluid.io`, version `v1alpha1`.

This is a library — no `cmd/` entry point. Other projects import and register the controller via `advancedstatefulset.Add(mgr)`.

## Module

Single Go module: `github.com/fluid-cloudnative/advanced-statefulset`

Dependencies are vendored in the root `vendor/` directory. Kubernetes libraries are pinned to v0.29.0 via `replace` directives in `go.mod`.

## Build & Development Commands

```bash
# Build all packages
make build
# or: go build ./...

# Generate CRDs and RBAC manifests (requires controller-gen)
make manifests

# Tidy dependencies
make tidy

# Vendor dependencies
make vendor

# Run tests
go test ./...

# Run a single test
go test ./pkg/workload/advancedstatefulset/ -run TestFunctionName

# Run the example controller
make run-example
```

## Architecture

### API Types (`api/workload/v1alpha1/`)

Single CRD: `AdvancedStatefulSet` — extends standard StatefulSet with:
- In-place pod updates (image, env, resources)
- Pod lifecycle hooks (PreDelete, InPlaceUpdate, PreNormal)
- Pod readiness gates
- Update priority/scatter strategies
- Reserve ordinals (skip specific pod indices)
- Volume claim update strategies
- Scale strategies with MaxUnavailable

### Controller (`pkg/workload/advancedstatefulset/`)

Uses `controller-runtime` and registers via `advancedstatefulset.Add(mgr)`. Key files:
- `statefulset_controller.go` — reconciler setup, watches, informer wiring
- `stateful_set_control.go` — core reconciliation logic (scale up/down, rolling update)
- `stateful_pod_control.go` — pod CRUD operations
- `stateful_update_utils.go` — update decision helpers
- `stateful_set_utils.go` — ordinal math, identity helpers
- `pvc_event_handler.go` — PVC event handling for volume claim updates

### Utility Packages (`pkg/workload/utils/`)

| Package | Purpose |
|---------|---------|
| `inplaceupdate` | In-place pod update logic (image, env-from-metadata, resources) |
| `lifecycle` | Pod lifecycle state machine (PreparingNormal -> Normal -> PreparingUpdate -> ...) |
| `expectations` | Scale/update/resourceVersion expectations to avoid stale-cache races |
| `updatesort` | Pod update ordering: priority-based and scatter-based sorting |
| `kubecontroller` | Controller ref manager and pod control (adopted from k8s upstream) |
| `controllerhistory` | ControllerRevision management |
| `revision` | Revision hash computation and comparison |
| `podreadiness` | KruisePodReady condition management |
| `specifieddelete` | Targeted pod deletion by annotation |
| `containermeta` | Container env hash computation for change detection |
| `discovery` | GVK discovery check (skip controller if CRD not installed) |

## Key Conventions

- Go 1.21 minimum
- Kubernetes v0.29.0, controller-runtime v0.17.0
- Code generation: `controller-gen` for CRDs (`config/crd/`) and RBAC (`config/rbac/`)
- DeepCopy: generated via `+kubebuilder:object:generate=true` markers, output in `zz_generated.deepcopy.go`
- Annotations/labels use `workload.fluid.io/` prefix (and `lifecycle.workload.fluid.io/` for lifecycle state)
