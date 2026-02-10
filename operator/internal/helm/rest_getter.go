package helm

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// restGetter implements genericclioptions.RESTClientGetter so Helm can use the in-cluster REST config.
type restGetter struct {
	config *rest.Config
}

func NewRESTGetter(cfg *rest.Config) genericclioptions.RESTClientGetter {
	return &restGetter{config: cfg}
}

func (r *restGetter) ToRESTConfig() (*rest.Config, error) {
	// return a shallow copy to be safe
	c := rest.CopyConfig(r.config)
	return c, nil
}

// ToDiscoveryClient returns a cached discovery client
func (r *restGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(r.config)
	if err != nil {
		return nil, err
	}
	return memory.NewMemCacheClient(dc), nil
}

// ToRESTMapper returns a RESTMapper backed by discovery with caching
func (r *restGetter) ToRESTMapper() (meta.RESTMapper, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(r.config)
	if err != nil {
		return nil, err
	}
	return restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc)), nil
}

// ToRawKubeConfigLoader returns a default (empty) kubeconfig loader.
// Helm/cli libraries sometimes call this; providing a default loader avoids panics.
// The loader is rarely used when we provide a REST config, but implement it to satisfy interface.
func (r *restGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return clientcmd.NewDefaultClientConfig(clientcmdapi.Config{}, &clientcmd.ConfigOverrides{})
}
