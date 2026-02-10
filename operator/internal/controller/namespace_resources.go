package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ensureQuota creates or updates a ResourceQuota
func (r *StoreReconciler) ensureQuota(ctx context.Context, namespace, plan string) error {
	logger := ctrl.LoggerFrom(ctx)
	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "store-resource-quota",
			Namespace: namespace,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				"requests.cpu":    resource.MustParse("1"),
				"requests.memory": resource.MustParse("1024Mi"), // Increased
				"limits.cpu":      resource.MustParse("2"),
				"limits.memory":   resource.MustParse("2048Mi"), // Increased to 2GB
				"pods":            resource.MustParse("15"),
			},
		},
	}

	var existing corev1.ResourceQuota
	if err := r.Get(ctx, clientObjectKey(namespace, "store-resource-quota"), &existing); err == nil {
		existing.Spec = quota.Spec
		if err := r.Update(ctx, &existing); err != nil {
			logger.Error(err, "updating ResourceQuota")
			return err
		}
		return nil
	}

	if err := r.Create(ctx, quota); err != nil {
		logger.Error(err, "creating ResourceQuota")
		return err
	}
	return nil
}

// ensureLimitRange creates defaults for containers
func (r *StoreReconciler) ensureLimitRange(ctx context.Context, namespace string) error {
	logger := ctrl.LoggerFrom(ctx)

	limit := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "store-limit-range",
			Namespace: namespace,
		},
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{
				{
					Type: corev1.LimitTypeContainer,
					Default: corev1.ResourceList{
						"cpu":    resource.MustParse("500m"),
						"memory": resource.MustParse("512Mi"), // Increased from 256Mi
					},
					DefaultRequest: corev1.ResourceList{
						"cpu":    resource.MustParse("100m"),
						"memory": resource.MustParse("256Mi"),
					},
				},
			},
		},
	}

	var existing corev1.LimitRange
	if err := r.Get(ctx, clientObjectKey(namespace, "store-limit-range"), &existing); err == nil {
		existing.Spec = limit.Spec
		if err := r.Update(ctx, &existing); err != nil {
			logger.Error(err, "updating LimitRange")
			return err
		}
		return nil
	}

	if err := r.Create(ctx, limit); err != nil {
		logger.Error(err, "creating LimitRange")
		return err
	}
	return nil
}

// ensureNetworkPolicy creates a default-deny policy
func (r *StoreReconciler) ensureNetworkPolicy(ctx context.Context, namespace string) error {
	logger := ctrl.LoggerFrom(ctx)
	name := "store-default-deny"

	np := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{}, // Selects ALL pods
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
				netv1.PolicyTypeEgress,
			},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					// Allow traffic from Ingress Controller
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": "ingress-nginx",
								},
							},
						},
						{
							// Also allow traffic from within the same namespace
							PodSelector: &metav1.LabelSelector{},
						},
					},
				},
			},
			Egress: []netv1.NetworkPolicyEgressRule{
				{
					// Allow all egress (for DNS, etc)
					To: []netv1.NetworkPolicyPeer{},
				},
			},
		},
	}

	var existing netv1.NetworkPolicy
	if err := r.Get(ctx, clientObjectKey(namespace, name), &existing); err == nil {
		existing.Spec = np.Spec
		if err := r.Update(ctx, &existing); err != nil {
			logger.Error(err, "updating NetworkPolicy")
			return err
		}
		return nil
	}

	if err := r.Create(ctx, np); err != nil {
		logger.Error(err, "creating NetworkPolicy")
		return err
	}
	return nil
}

func clientObjectKey(namespace, name string) client.ObjectKey {
	return client.ObjectKey{Namespace: namespace, Name: name}
}
