package controller

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Jovial-Kanwadia/store-operator/internal/helm"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1alpha1 "github.com/Jovial-Kanwadia/store-operator/api/v1alpha1"
)

const storeFinalizer = "infra.store.io/finalizer"

// StoreReconciler reconciles a Store object
type StoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

			// A. Uninstall Helm Release
			if err := helm.UninstallRelease(ctrl.GetConfigOrDie(), releaseName, nsName); err != nil {
				logger.Error(err, "Helm uninstall failed")
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
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
			if err := r.Get(ctx, types.NamespacedName{Name: nsName}, &ns); err == nil {
				if err := r.Delete(ctx, &ns); err != nil {
					return ctrl.Result{}, err
				}
				// Wait for namespace to actually vanish
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}

			// D. Remove Finalizer
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
		// Fallback for local dev
		chartPath = "/Users/zora/programming/store-platform/operator/charts/wordpress"
	}

	// E. Prepare Values
	values := map[string]interface{}{
		"wordpressBlogName": store.Name,
		"service":           map[string]interface{}{"type": "ClusterIP"},
		"ingress": map[string]interface{}{
			"enabled":          true,
			"ingressClassName": "nginx",
			"hostname":         fmt.Sprintf("%s.127.0.0.1.nip.io", store.Name),
		},
		"volumePermissions": map[string]interface{}{"enabled": false},

		// 1. Disable Persistence (Critical for Kind/Local speed)
		"persistence": map[string]interface{}{
			"enabled": false,
		},
		"mariadb": map[string]interface{}{
			"primary": map[string]interface{}{
				"persistence": map[string]interface{}{
					"enabled": false,
				},
			},
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
		r.Status().Update(ctx, &store)
	}

	if err := helm.InstallOrUpgrade(ctx, ctrl.GetConfigOrDie(), releaseName, nsName, chartPath, values); err != nil {
		logger.Error(err, "Helm install failed")
		store.Status.Phase = "Failed"
		r.Status().Update(ctx, &store)
		return ctrl.Result{RequeueAfter: 20 * time.Second}, nil
	}

	// G. Verify Readiness (Probe URL)
	storeURL := fmt.Sprintf("http://%s.127.0.0.1.nip.io", store.Name)
	if err := r.probeURL(ctx, storeURL); err != nil {
		logger.Info("Waiting for Store URL...", "url", storeURL)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// H. Success!
	if store.Status.Phase != "Ready" {
		store.Status.Phase = "Ready"
		store.Status.URL = storeURL
		r.Status().Update(ctx, &store)
	}

	return ctrl.Result{}, nil
}

// probeURL tries to hit the URL. Returns nil if 200 OK.
func (r *StoreReconciler) probeURL(ctx context.Context, url string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil
	}
	return fmt.Errorf("status code %d", resp.StatusCode)
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
