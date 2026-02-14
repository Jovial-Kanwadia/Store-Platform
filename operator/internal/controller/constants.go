package controller

// Finalizer name
const storeFinalizer = "infra.store.io/finalizer"

// Store status phases
const (
	PhaseProvisioning = "Provisioning"
	PhaseReady        = "Ready"
	PhaseFailed       = "Failed"
)

// Store status reasons
const (
	ReasonProvisioning   = "Provisioning"
	ReasonHelmError      = "HelmError"
	ReasonWaitingForPods = "WaitingForPods"
)

// Kubernetes resource names
const (
	ResourceQuotaName = "store-resource-quota"
	LimitRangeName    = "store-limit-range"
	NetworkPolicyName = "store-default-deny"
)

// Namespace naming
const (
	StoreNamespacePrefix = "store-"
)

// Secret keys
const (
	SecretKeyMariaDBRoot = "mariadb-root-password"
	SecretKeyMariaDBUser = "mariadb-user-password"
	SecretKeyWordPress   = "wordpress-password"
)

// WordPress Helm chart labels
const (
	WordPressAppLabel = "app.kubernetes.io/name"
	WordPressAppValue = "wordpress"
)

// Password generation lengths
const (
	MariaDBRootPasswordLength = 20
	MariaDBUserPasswordLength = 20
	WordPressPasswordLength   = 16
)

// Event reasons
const (
	EventReasonDeleteFailed = "DeleteFailed"
	EventReasonProvisioning = "Provisioning"
	EventReasonFailed       = "Failed"
	EventReasonReady        = "Ready"
)

// Helm values keys (for documentation and consistency)
const (
	HelmKeyWordPressBlogName = "wordpressBlogName"
	HelmKeyService           = "service"
	HelmKeyVolumePermissions = "volumePermissions"
	HelmKeyWordPressPassword = "wordpressPassword"
	HelmKeyIngress           = "ingress"
	HelmKeyMariaDB           = "mariadb"
	HelmKeyPersistence       = "persistence"
	HelmKeyLivenessProbe     = "livenessProbe"
	HelmKeyReadinessProbe    = "readinessProbe"
)
