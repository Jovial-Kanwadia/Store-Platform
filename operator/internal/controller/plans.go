package controller

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// PlanSpec defines resource limits and defaults for a store plan
type PlanSpec struct {
	// ResourceQuota limits
	RequestsCPU    resource.Quantity
	RequestsMemory resource.Quantity
	LimitsCPU      resource.Quantity
	LimitsMemory   resource.Quantity
	MaxPods        resource.Quantity

	// LimitRange defaults
	DefaultCPU           resource.Quantity
	DefaultMemory        resource.Quantity
	DefaultRequestCPU    resource.Quantity
	DefaultRequestMemory resource.Quantity
}

// SupportedPlans defines the resource specifications for each plan tier
var SupportedPlans = map[string]PlanSpec{
	"small": {
		RequestsCPU:    resource.MustParse("500m"),
		RequestsMemory: resource.MustParse("512Mi"),
		LimitsCPU:      resource.MustParse("1"),
		LimitsMemory:   resource.MustParse("1Gi"),
		MaxPods:        resource.MustParse("10"),

		DefaultCPU:           resource.MustParse("200m"),
		DefaultMemory:        resource.MustParse("256Mi"),
		DefaultRequestCPU:    resource.MustParse("50m"),
		DefaultRequestMemory: resource.MustParse("128Mi"),
	},
	"medium": {
		RequestsCPU:    resource.MustParse("1"),
		RequestsMemory: resource.MustParse("1Gi"),
		LimitsCPU:      resource.MustParse("2"),
		LimitsMemory:   resource.MustParse("2Gi"),
		MaxPods:        resource.MustParse("15"),

		DefaultCPU:           resource.MustParse("500m"),
		DefaultMemory:        resource.MustParse("512Mi"),
		DefaultRequestCPU:    resource.MustParse("100m"),
		DefaultRequestMemory: resource.MustParse("256Mi"),
	},
	"large": {
		RequestsCPU:    resource.MustParse("2"),
		RequestsMemory: resource.MustParse("2Gi"),
		LimitsCPU:      resource.MustParse("4"),
		LimitsMemory:   resource.MustParse("4Gi"),
		MaxPods:        resource.MustParse("20"),

		DefaultCPU:           resource.MustParse("1"),
		DefaultMemory:        resource.MustParse("1Gi"),
		DefaultRequestCPU:    resource.MustParse("200m"),
		DefaultRequestMemory: resource.MustParse("512Mi"),
	},
}

// GetPlanSpec returns the resource spec for a plan, defaulting to "small" if invalid
func GetPlanSpec(plan string) PlanSpec {
	if spec, ok := SupportedPlans[plan]; ok {
		return spec
	}
	// Default to small plan if invalid plan specified
	return SupportedPlans["small"]
}

// IsValidPlan checks if a plan name is supported
func IsValidPlan(plan string) bool {
	_, ok := SupportedPlans[plan]
	return ok
}
