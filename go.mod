module github.com/fluid-cloudnative/common

go 1.21

require (
	github.com/fluid-cloudnative/common/api v0.0.0
	github.com/fluid-cloudnative/common/pkg/workload v0.0.0
)

replace (
	github.com/fluid-cloudnative/common/api => ./api
	github.com/fluid-cloudnative/common/pkg/workload => ./pkg/workload
)
