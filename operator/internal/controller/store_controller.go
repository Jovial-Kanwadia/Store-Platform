/*
Copyright 2026.

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

package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1alpha1 "github.com/Jovial-Kanwadia/store-operator/api/v1alpha1"
)

// StoreReconciler reconciles a Store object
type StoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infra.store.io,resources=stores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infra.store.io,resources=stores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infra.store.io,resources=stores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Store object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.1/pkg/reconcile
func (r *StoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	// 1. Fetch the Store object
	var store infrav1alpha1.Store
	if err := r.Get(ctx, req.NamespacedName, &store); err != nil {
		if errors.IsNotFound(err) {
			// Store was deleted
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 2. Desired namespace name
	nsName := fmt.Sprintf("store-%s", store.Name)

	// 3. Check if namespace already exists
	var ns corev1.Namespace
	err := r.Get(ctx, types.NamespacedName{Name: nsName}, &ns)

	if err != nil && errors.IsNotFound(err) {
		// 4. Namespace does not exist → create it
		logger.Info("Creating namespace for store", "namespace", nsName)

		ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: nsName,
			},
		}

		if err := r.Create(ctx, &ns); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 5. Update Store status to Provisioning
	if store.Status.Phase != "Provisioning" {
		store.Status.Phase = "Provisioning"
		if err := r.Status().Update(ctx, &store); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1alpha1.Store{}).
		Named("store").
		Complete(r)
}
