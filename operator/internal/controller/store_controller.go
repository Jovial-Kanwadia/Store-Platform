package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"os"
	"time"

	"github.com/Jovial-Kanwadia/store-operator/internal/helm"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1alpha1 "github.com/Jovial-Kanwadia/store-operator/api/v1alpha1"
)

const storeFinalizer = "infra.store.io/finalizer"

// StoreReconciler reconciles a Store object
type StoreReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=infra.store.io,resources=stores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infra.store.io,resources=stores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infra.store.io,resources=stores/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=pods;services;events;persistentvolumeclaims;secrets;configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="networking.k8s.io",resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=resourcequotas;limitranges,verbs=get;list;watch;create;update;patch;delete

func (r *StoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var store infrav1alpha1.Store
	if err := r.Get(ctx, req.NamespacedName, &store); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	nsName := fmt.Sprintf("store-%s", store.Name)
	releaseName := store.Name

	// 1. DELETE LOGIC
	if !store.DeletionTimestamp.IsZero() {
		if containsString(store.Finalizers, storeFinalizer) {
			logger.Info("Deleting Store resources...", "store", store.Name)
			storeDeletionTotal.Inc()

			// A. Uninstall Helm Release
			// A. Uninstall Helm Release
			if err := helm.UninstallRelease(ctrl.GetConfigOrDie(), releaseName, nsName); err != nil {
				// Ignore "not found" errors to prevent getting stuck
				if !strings.Contains(err.Error(), "not found") {
					logger.Error(err, "Helm uninstall failed")
					r.Recorder.Eventf(&store, corev1.EventTypeWarning, "DeleteFailed", "Helm uninstall failed: %v", err)
					return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
				}
			}

			// B. Delete PVCs (Clean up storage)
			var pvcList corev1.PersistentVolumeClaimList
			if err := r.List(ctx, &pvcList, &client.ListOptions{Namespace: nsName}); err == nil {
				for _, pvc := range pvcList.Items {
					_ = r.Delete(ctx, &pvc)
				}
			}

			// C. Delete Namespace
			var ns corev1.Namespace
			err := r.Get(ctx, types.NamespacedName{Name: nsName}, &ns)
			if err == nil {
				// Namespace exists - DELETE IT
				if ns.Status.Phase != corev1.NamespaceTerminating {
					logger.Info("Deleting Namespace", "namespace", nsName)
					if err := r.Delete(ctx, &ns); err != nil {
						return ctrl.Result{}, err
					}
				}
				// Wait for namespace to actually vanish
				logger.Info("Waiting for Namespace to terminate...", "namespace", nsName)
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			} else if !apierrors.IsNotFound(err) {
				// Real error (e.g. connection refused) - DO NOT remove finalizer yet
				logger.Error(err, "Failed to check namespace existence")
				return ctrl.Result{}, err
			}

			// D. Remove Finalizer (Only reachable if Namespace is NotFound)
			store.Finalizers = removeString(store.Finalizers, storeFinalizer)
			if err := r.Update(ctx, &store); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// 2. CREATE/UPDATE LOGIC
	// A. Add Finalizer
	if !containsString(store.Finalizers, storeFinalizer) {
		store.Finalizers = append(store.Finalizers, storeFinalizer)
		if err := r.Update(ctx, &store); err != nil {
			return ctrl.Result{}, err
		}
	}

	// B. Ensure Namespace
	var ns corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: nsName}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			ns = corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}
			if err := r.Create(ctx, &ns); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
		}
		return ctrl.Result{}, err
	}

	// NEW: Manage Credentials
	creds, err := r.ReconcileCredentials(ctx, &store)
	if err != nil {
		return ctrl.Result{}, err
	}

	// C. Apply Guardrails (Quota, Limits, NetPol)
	if err := r.ensureQuota(ctx, nsName, store.Spec.Plan); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.ensureLimitRange(ctx, nsName); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.ensureNetworkPolicy(ctx, nsName); err != nil {
		return ctrl.Result{}, err
	}

	// D. Determine Chart Path
	chartPath := os.Getenv("WORDPRESS_CHART_PATH")
	if chartPath == "" {
		chartPath = "../charts/engine-woo"
	}

	baseDomain := os.Getenv("BASE_DOMAIN")
	if baseDomain == "" {
		baseDomain = "127.0.0.1.nip.io"
	}

	// E. Prepare Values
	values := map[string]interface{}{
		"wordpressBlogName": store.Name,
		"service":           map[string]interface{}{"type": "ClusterIP"},
		"volumePermissions": map[string]interface{}{"enabled": false},

		// Inject Credentials & Networking
		"wordpressPassword": creds["wordpress-password"],
		"ingress": map[string]interface{}{
			"enabled":          true,
			"ingressClassName": "nginx",
			"hostname":         fmt.Sprintf("%s.%s", store.Name, baseDomain),
		},
		"mariadb": map[string]interface{}{
			"auth": map[string]interface{}{
				"rootPassword": creds["mariadb-root-password"],
				"password":     creds["mariadb-user-password"],
			},
			"primary": map[string]interface{}{
				"persistence": map[string]interface{}{
					"enabled": false,
				},
			},
		},

		// 1. Disable Persistence (Critical for Kind/Local speed)
		"persistence": map[string]interface{}{
			"enabled": false,
		},

		// 2. Relax Probes (Critical for slow laptops)
		// Give WordPress 2 minutes to start before Kubernetes kills it
		"livenessProbe": map[string]interface{}{
			"initialDelaySeconds": 120,
			"periodSeconds":       20,
		},
		"readinessProbe": map[string]interface{}{
			"initialDelaySeconds": 60,
			"periodSeconds":       10,
		},
	}

	// F. Install/Upgrade Helm
	if store.Status.Phase == "" {
		store.Status.Phase = "Provisioning"
		store.Status.Message = "Started provisioning store"
		store.Status.Reason = "Provisioning"
		if err := r.Status().Update(ctx, &store); err != nil {
			logger.Error(err, "unable to update Store status")
			return ctrl.Result{}, err
		}

		r.Recorder.Event(&store, corev1.EventTypeNormal, "Provisioning", fmt.Sprintf("Started provisioning store %s", store.Name))
		storeCreatedTotal.Inc()
	}

	provisionStart := store.CreationTimestamp.Time

	// CHECK IDEMPOTENCY: Only run Helm if Spec changed or not ready
	if store.Generation != store.Status.ObservedGeneration || store.Status.Phase != "Ready" {
		if err := helm.InstallOrUpgrade(ctx, ctrl.GetConfigOrDie(), releaseName, nsName, chartPath, values); err != nil {
			logger.Error(err, "Helm install failed")
			store.Status.Phase = "Failed"
			store.Status.Message = fmt.Sprintf("Helm install failed: %v", err)
			store.Status.Reason = "HelmError"
			if err := r.Status().Update(ctx, &store); err != nil {
				logger.Error(err, "unable to update Store status")
				return ctrl.Result{}, err
			}

			r.Recorder.Eventf(&store, corev1.EventTypeWarning, "Failed", "Installation failed: %v", err)
			return ctrl.Result{RequeueAfter: 20 * time.Second}, nil
		}
		// Update ObservedGeneration after successful Helm run
		store.Status.ObservedGeneration = store.Generation
		// We don't update status here yet, we wait until final success to save API calls
	}

	// G. Verify Readiness (Check if Pod is Ready)
	// We use the Kubernetes API instead of HTTP probing because probing internal
	// cluster IPs from a local operator (outside the cluster) is flaky/impossible.
	if !r.isPodReady(ctx, nsName) {
		logger.Info("Waiting for Pods to be Ready...", "namespace", nsName)
		store.Status.Message = "Waiting for pods to become ready..."
		store.Status.Reason = "WaitingForPods"
		if err := r.Status().Update(ctx, &store); err != nil {
			logger.Error(err, "unable to update Store status")
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// H. Success!
	if store.Status.Phase != "Ready" {
		store.Status.Phase = "Ready"
		storeURL := fmt.Sprintf("http://%s.%s", store.Name, baseDomain)
		store.Status.URL = storeURL
		store.Status.Message = ""
		store.Status.Reason = ""
		if err := r.Status().Update(ctx, &store); err != nil {
			logger.Error(err, "unable to update Store status")
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&store, corev1.EventTypeNormal, "Ready", "Store is ready at URL %s", storeURL)
		storeProvisioningSeconds.Observe(time.Since(provisionStart).Seconds())
	}

	return ctrl.Result{}, nil
}

// isPodReady checks if there is at least one running and ready Pod for the WordPress app
func (r *StoreReconciler) isPodReady(ctx context.Context, namespace string) bool {
	var podList corev1.PodList
	// Bitnami WordPress charts use this label by default
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{"app.kubernetes.io/name": "wordpress"},
	}

	if err := r.List(ctx, &podList, opts...); err != nil {
		return false
	}

	if len(podList.Items) == 0 {
		return false
	}

	for _, pod := range podList.Items {
		// Check if Pod is Running
		if pod.Status.Phase == corev1.PodRunning {
			// Check if it's Ready
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
					return true
				}
			}
		}
	}
	return false
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	var result []string
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

func (r *StoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1alpha1.Store{}).
		Named("store").
		Complete(r)
}

// Generate a random secure password
func generatePassword(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "default-secure-password-change-me" // Fallback (should rarely happen)
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// ReconcileCredentials ensures a secret exists with stable passwords
func (r *StoreReconciler) ReconcileCredentials(ctx context.Context, store *infrav1alpha1.Store) (map[string]string, error) {
	secretName := store.Name + "-creds"
	secret := &corev1.Secret{}

	// 1. Try to fetch existing secret
	err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: store.Namespace}, secret)

	if err != nil && apierrors.IsNotFound(err) {
		// 2. Secret doesn't exist? Create it!
		log.Log.Info("Generating new credentials for store", "store", store.Name)

		creds := map[string]string{
			"mariadb-root-password": generatePassword(20),
			"mariadb-user-password": generatePassword(20),
			"wordpress-password":    generatePassword(16),
		}

		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: store.Namespace, // Store it in the SAME namespace as the store
			},
			StringData: creds,
		}

		// Set OwnerReference so deleting the Store deletes the Secret
		if err := controllerutil.SetControllerReference(store, secret, r.Scheme); err != nil {
			return nil, err
		}

		if err := r.Create(ctx, secret); err != nil {
			return nil, err
		}
		return creds, nil
	} else if err != nil {
		return nil, err
	}

	// 3. Secret exists? Return existing values
	existingCreds := make(map[string]string)
	for k, v := range secret.Data {
		existingCreds[k] = string(v)
	}
	return existingCreds, nil
}
