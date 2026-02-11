package k8s

import (
	"context"
	"fmt"

	"github.com/Jovial-Kanwadia/store-platform/backend/internal/domain"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var storeGVR = schema.GroupVersionResource{
	Group:    "infra.store.io",
	Version:  "v1alpha1",
	Resource: "stores",
}

type Client struct {
	dynamicClient dynamic.Interface
}

func NewClient(kubeconfigPath string) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build kubernetes config: %w", err)
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Client{
		dynamicClient: dynClient,
	}, nil
}

func (c *Client) Create(ctx context.Context, s domain.Store) error {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "infra.store.io/v1alpha1",
			"kind":       "Store",
			"metadata": map[string]interface{}{
				"name":      s.Name,
				"namespace": s.Namespace,
			},
			"spec": map[string]interface{}{
				"engine": s.Engine,
				"plan":   s.Plan,
			},
		},
	}

	_, err := c.dynamicClient.Resource(storeGVR).Namespace(s.Namespace).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	return nil
}

func (c *Client) List(ctx context.Context, namespace string) ([]domain.Store, error) {
	var list *unstructured.UnstructuredList
	var err error

	if namespace != "" {
		list, err = c.dynamicClient.Resource(storeGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	} else {
		list, err = c.dynamicClient.Resource(storeGVR).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list stores: %w", err)
	}

	stores := make([]domain.Store, 0, len(list.Items))
	for _, item := range list.Items {
		store, err := unstructuredToStore(&item)
		if err != nil {
			continue
		}
		stores = append(stores, *store)
	}

	return stores, nil
}

func (c *Client) Get(ctx context.Context, name, namespace string) (*domain.Store, error) {
	obj, err := c.dynamicClient.Resource(storeGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get store: %w", err)
	}

	return unstructuredToStore(obj)
}

func (c *Client) Delete(ctx context.Context, name, namespace string) error {
	err := c.dynamicClient.Resource(storeGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete store: %w", err)
	}

	return nil
}

func unstructuredToStore(obj *unstructured.Unstructured) (*domain.Store, error) {
	name := obj.GetName()
	namespace := obj.GetNamespace()
	createdAt := obj.GetCreationTimestamp().Time

	// 1. Extract Spec
	spec, found, err := unstructured.NestedMap(obj.Object, "spec")
	if err != nil || !found {
		spec = make(map[string]interface{})
	}

	// 2. Extract Fields safely
	statusMap, _, _ := unstructured.NestedMap(obj.Object, "status")
	phase, _, _ := unstructured.NestedString(statusMap, "phase")
	url, _, _ := unstructured.NestedString(statusMap, "url")
	
	// FIX: Read Engine and Plan so they appear in the API response
	engine, _, _ := unstructured.NestedString(spec, "engine")
	plan, _, _ := unstructured.NestedString(spec, "plan")

	return &domain.Store{
		Name:      name,
		Namespace: namespace,
		// Map them to the flat struct
		Engine:    engine,
		Plan:      plan,
		Status:    phase, // "phase" from K8s maps to "status" in our JSON
		URL:       url,
		CreatedAt: createdAt,
	}, nil
}