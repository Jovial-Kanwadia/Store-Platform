package controller

import (
	"context"
	"fmt"
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
// +kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services;pods;persistentvolumeclaims;secrets;configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *StoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var store infrav1alpha1.Store
	if err := r.Get(ctx, req.NamespacedName, &store); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	nsName := fmt.Sprintf("store-%s", store.Name)

	// --- DELETE FLOW ---
	if !store.DeletionTimestamp.IsZero() {
		if containsString(store.Finalizers, storeFinalizer) {
			logger.Info("handling deletion", "store", store.Name)
			var ns corev1.Namespace
			err := r.Get(ctx, types.NamespacedName{Name: nsName}, &ns)
			if err == nil {
				// trigger deletion (idempotent)
				if delErr := r.Delete(ctx, &ns); delErr != nil && !apierrors.IsNotFound(delErr) {
					return ctrl.Result{}, delErr
				}
				// wait for namespace to disappear
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}
			if !apierrors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			// namespace gone → remove finalizer
			store.Finalizers = removeString(store.Finalizers, storeFinalizer)
			if err := r.Update(ctx, &store); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// --- CREATE / UPDATE FLOW ---

	// ensure finalizer
	if !containsString(store.Finalizers, storeFinalizer) {
		store.Finalizers = append(store.Finalizers, storeFinalizer)
		if err := r.Update(ctx, &store); err != nil {
			return ctrl.Result{}, err
		}
		// requeue to see new finalizer
		return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
	}

	// ensure namespace exists
	var ns corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: nsName}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("creating namespace", "namespace", nsName)
			ns = corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: nsName,
				},
			}
			if err := r.Create(ctx, &ns); err != nil {
				return ctrl.Result{}, err
			}
			// requeue to allow namespace objects to be created
			return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
		}
		return ctrl.Result{}, err
	}

	// chart path via env var (required)
	chartPath := os.Getenv("WORDPRESS_CHART_PATH")
	if chartPath == "" {
		return ctrl.Result{}, fmt.Errorf("WORDPRESS_CHART_PATH not set in operator environment")
	}

	releaseName := store.Name
	values := map[string]interface{}{
		"wordpressBlogName": store.Name,
		// Fix the Networking Issue
		"service": map[string]interface{}{
			"type": "ClusterIP",
		},
		// Enable Ingress (So you can actually access it via nip.io)
		"ingress": map[string]interface{}{
			"enabled":          true,
			"ingressClassName": "nginx",
			"hostname":         fmt.Sprintf("%s.127.0.0.1.nip.io", store.Name),
		},
	}

	// set status to Provisioning if not set
	if store.Status.Phase == "" || store.Status.Phase == "Failed" {
		store.Status.Phase = "Provisioning"
		if err := r.Status().Update(ctx, &store); err != nil {
			return ctrl.Result{}, err
		}
		// let next reconcile attempt run helm
		return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
	}

	// Perform Helm install/upgrade using in-cluster config
	if err := helm.InstallOrUpgrade(ctx, ctrl.GetConfigOrDie(), releaseName, nsName, chartPath, values); err != nil {
		logger.Error(err, "helm install/upgrade failed", "release", releaseName, "namespace", nsName)
		// set status to Failed and include short message (avoid long errors in status)
		store.Status.Phase = "Failed"
		if err2 := r.Status().Update(ctx, &store); err2 != nil {
			logger.Error(err2, "failed to update status after helm failure")
		}
		// requeue after a backoff to allow transient issues to recover
		return ctrl.Result{RequeueAfter: 20 * time.Second}, nil
	}

	// success → Ready
	if store.Status.Phase != "Ready" {
		store.Status.Phase = "Ready"
		if err := r.Status().Update(ctx, &store); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
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
