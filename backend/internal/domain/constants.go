package domain

// Store status phases — must match operator constants in
// operator/internal/controller/constants.go
const (
	StatusPending      = "Pending"
	StatusProvisioning = "Provisioning"
	StatusReady        = "Ready"
	StatusFailed       = "Failed"
)

// Supported plans — must match operator plans in
// operator/internal/controller/plans.go
const (
	PlanSmall  = "small"
	PlanMedium = "medium"
	PlanLarge  = "large"
)

// AllowedPlans is the set of valid plan names.
var AllowedPlans = map[string]bool{
	PlanSmall:  true,
	PlanMedium: true,
	PlanLarge:  true,
}

// Supported engines
const (
	EngineWoo = "woo"
)

// AllowedEngines is the set of valid engine names.
var AllowedEngines = map[string]bool{
	EngineWoo: true,
}

// Default values
const (
	DefaultNamespace = "default"
)

// Validation limits
const (
	MaxStoreNameLength = 63
)

// CRD metadata — must match the operator CRD definition in
// operator/api/v1alpha1/store_types.go
const (
	CRDGroup      = "infra.store.io"
	CRDVersion    = "v1alpha1"
	CRDResource   = "stores"
	CRDKind       = "Store"
	CRDAPIVersion = CRDGroup + "/" + CRDVersion
)
