package helm

import (
	"fmt"
	"os"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/rest"
)

// UninstallRelease uninstalls a Helm release from namespace using the provided restConfig.
func UninstallRelease(restConfig *rest.Config, releaseName, namespace string) error {
	settings := cli.New()
	settings.SetNamespace(namespace)

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(
		NewRESTGetter(restConfig),
		namespace,
		os.Getenv("HELM_DRIVER"),
		func(string, ...interface{}) {},
	); err != nil {
		return err
	}

	uninstall := action.NewUninstall(actionConfig)
	_, err := uninstall.Run(releaseName)
	if err != nil {
		// If Helm says "release: not found", consider it already uninstalled (Success)
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		// Otherwise, error
		return fmt.Errorf("helm uninstall failed: %w", err)
	}
	return nil
}

// GetReleaseRevision returns the latest revision number for a release, or -1 if none.
func GetReleaseRevision(restConfig *rest.Config, releaseName, namespace string) (int, error) {
	settings := cli.New()
	settings.SetNamespace(namespace)

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(
		NewRESTGetter(restConfig),
		namespace,
		os.Getenv("HELM_DRIVER"),
		func(string, ...interface{}) {},
	); err != nil {
		return -1, err
	}

	hist := action.NewHistory(actionConfig)
	hist.Max = 1
	rel, err := hist.Run(releaseName)
	if err != nil {
		return -1, err
	}
	if len(rel) == 0 {
		return -1, nil
	}
	return rel[0].Version, nil
}
