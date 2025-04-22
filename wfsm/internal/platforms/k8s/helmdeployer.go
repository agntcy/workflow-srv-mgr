package k8s

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// helmDeployer helm based deployer implementation
type helmDeployer struct {
}

type HelmDeploymentService interface {
	DeployChart(ctx context.Context, releaseName string, chartUrl string, namespace string, chartValuesYaml []byte) error
	UnDeployChart(ctx context.Context, releaseName string, namespace string) error
}

func NewHelmDeployer() HelmDeploymentService {
	return helmDeployer{}
}

func (h helmDeployer) isUpgrade(cfg action.Configuration, releaseName string) bool {
	history, err := cfg.Releases.History(releaseName)
	if err != nil || len(history) < 1 {
		return false
	}
	return true
}

// Loads Chart from remote repository by chartURL or legacy filesystem store
func (h helmDeployer) pullChart(chartPathOpts *action.ChartPathOptions, chartUrl string) (*chart.Chart, error) {
	envSettings := cli.New()
	chartPath, err := chartPathOpts.LocateChart(chartUrl, envSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}
	chartRequested, err := loader.Load(chartPath)
	return chartRequested, err
}

func (h helmDeployer) DeployChart(ctx context.Context, releaseName string, chartUrl string, namespace string, chartValuesYaml []byte) error {
	log := zerolog.Ctx(ctx)
	log.Info().Str("chartURL", chartUrl).Msg("Deploying chart")

	// Initialize the action configuration
	helmActionConfiguration, err := h.getActionConfiguration(namespace)
	if err != nil {
		return fmt.Errorf("failed to initialize the action configuration: %w", err)
	}

	if h.isUpgrade(helmActionConfiguration, releaseName) {
		upgradeAction := action.NewUpgrade(&helmActionConfiguration)
		upgradeAction.Namespace = namespace

		chartValues, err := h.convertValuesToMap(chartValuesYaml)
		if err != nil {
			return err
		}

		chartRequested, err := h.pullChart(&upgradeAction.ChartPathOptions, chartUrl)
		if err != nil {
			return fmt.Errorf("failed to load chart: %w", err)
		}

		release, err := upgradeAction.Run(releaseName, chartRequested, chartValues)
		if err != nil {
			return fmt.Errorf("failed to upgrade chart: %w", err)
		}
		log.Info().Str("release", release.Name).Msg("Chart successfully upgraded")
	} else {
		installAction := action.NewInstall(&helmActionConfiguration)
		installAction.ReleaseName = releaseName
		installAction.Namespace = namespace
		installAction.CreateNamespace = true

		chartRequested, err := h.pullChart(&installAction.ChartPathOptions, chartUrl)
		if err != nil {
			return fmt.Errorf("failed to load chart: %w", err)
		}

		chartValues, err := h.convertValuesToMap(chartValuesYaml)
		if err != nil {
			return fmt.Errorf("failed to prepare configuration values: %w", err)
		}

		release, err := installAction.Run(chartRequested, chartValues)
		if err != nil {
			return fmt.Errorf("failed to install chart: %w", err)
		}
		log.Info().Str("release", release.Name).Msg("Chart successfully deployed")
	}

	return nil
}

func (h helmDeployer) convertValuesToMap(chartValuesYaml []byte) (map[string]interface{}, error) {
	chartValuesMap := make(map[string]interface{})
	if err := yaml.Unmarshal(chartValuesYaml, chartValuesMap); err != nil {
		return nil, fmt.Errorf("failed to decode chart values: %w", err)
	}
	return chartValuesMap, nil
}

func (h helmDeployer) UnDeployChart(ctx context.Context, releaseName string, namespace string) error {
	log := zerolog.Ctx(ctx)
	log.Info().Str("releaseName", releaseName).Msg("Uninstalling chart")

	helmActionConfiguration, err := h.getActionConfiguration(namespace)
	if err != nil {
		return fmt.Errorf("failed to uninstall chart: %v", err)
	}

	_, uninstallErr := action.NewUninstall(&helmActionConfiguration).Run(releaseName)
	if uninstallErr != nil {
		if _, err := helmActionConfiguration.Releases.History(releaseName); err != nil {
			log.Info().Str("releaseName", releaseName).Msg("release not found")
			return nil
		}
		return fmt.Errorf("failed to uninstall chart: %v", uninstallErr)
	}

	return nil
}

// getActionConfiguration assembles an "in-cluster" action configuration to be used for helm operations
func (h helmDeployer) getActionConfiguration(namespace string) (action.Configuration, error) {
	config := action.Configuration{}
	cfgFlags := genericclioptions.NewConfigFlags(true)
	cfgFlags.Namespace = &namespace
	err := config.Init(cfgFlags, namespace, "secret", func(format string, v ...interface{}) {})
	if err != nil {
		return action.Configuration{}, fmt.Errorf("failed to initialize the action configuration: %w", err)
	}
	return config, nil
}
