package helm

import (
	"context"
	"fmt"
	"os"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/rest"
)

func InstallOrUpgrade(
	ctx context.Context,
	restConfig *rest.Config,
	releaseName string,
	namespace string,
	chartPath string,
	values map[string]interface{},
) error {

	settings := cli.New()
	settings.SetNamespace(namespace)

	actionConfig := new(action.Configuration)
	// We add a logger here so you can see Helm output in your terminal
	if err := actionConfig.Init(
		NewRESTGetter(restConfig),
		namespace,
		os.Getenv("HELM_DRIVER"),
		func(format string, v ...interface{}) {
			fmt.Printf(format+"\n", v...)
		},
	); err != nil {
		return err
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	// 1. Check if the release already exists
	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(releaseName); err == nil {
		// Release exists -> UPGRADE
		fmt.Printf("Helm: Release %s exists, upgrading...\n", releaseName)
		upgrade := action.NewUpgrade(actionConfig)
		upgrade.Namespace = namespace
		upgrade.Wait = true
		upgrade.Timeout = 5 * time.Minute
		_, err := upgrade.Run(releaseName, chart, values)
		return err
	}

	// 2. Release does not exist -> INSTALL
	fmt.Printf("Helm: Release %s does not exist, installing...\n", releaseName)
	install := action.NewInstall(actionConfig)
	install.Namespace = namespace
	install.ReleaseName = releaseName
	install.Wait = true
	install.Timeout = 5 * time.Minute
	_, err = install.Run(chart, values)
	return err
}
