package deploy

import (
	"context"
	"os"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type AppConfig = map[string]interface{}

type DeploymentParams struct {
	DeployID             string
	SiteID               string
	ApplicationId        string
	ApplicationVersion   string
	AppName              string
	AppsNamespace        string
	IngressURL           string
	IngressHostname      string
	IstioGateway         string
	ServicePath          string
	ChartURL             string
	LogCollectionEnabled bool
	Config               AppConfig
}

type NetworkConfig struct {
	HTTPProxy  string
	HTTPSProxy string
	NoProxy    string
}

func NewNetworkConfigFromEnv() NetworkConfig {
	// uppercase vars are not considered here because they are already canonicalized to lc
	return NetworkConfig{
		HTTPProxy:  os.Getenv("http_proxy"),
		HTTPSProxy: os.Getenv("https_proxy"),
		NoProxy:    os.Getenv("no_proxy"),
	}
}

func (n NetworkConfig) Map() map[string]interface{} {
	return map[string]interface{}{
		"http_proxy":  n.HTTPProxy,
		"https_proxy": n.HTTPSProxy,
		"no_proxy":    n.NoProxy,
	}
}

// helmDeployer helm based deployer implementation
type helmDeployer struct {
	logger        logr.Logger
	networkConfig NetworkConfig
	ociConfig     edge.OCIConfig
}

type ConfigOpt func(*action.Configuration)

func registryConfigOpt(in *registry.Client) ConfigOpt {
	return func(config *action.Configuration) {
		config.RegistryClient = in
	}
}

func NewHelmDeployer(logger logr.Logger, networkConfig NetworkConfig, ociConfig edge.OCIConfig) AppDeploymentService {
	return helmDeployer{
		logger:        logger,
		networkConfig: networkConfig,
		ociConfig:     ociConfig,
	}
}

func (h helmDeployer) isUpgrade(cfg action.Configuration, releaseName string) bool {
	history, err := cfg.Releases.History(releaseName)
	if err != nil || len(history) < 1 {
		return false
	}
	return true
}

// Loads Chart from remote repository by chartURL or legacy filesystem store
func (h helmDeployer) PullChart(chartPathOpts *action.ChartPathOptions, deploySpec DeploymentParams) (*chart.Chart, error) {
	envSettings := cli.New()
	chartPath, err := chartPathOpts.LocateChart(deploySpec.ChartURL, envSettings)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to locate chart")
	}
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}
	chartRequested = h.deploymentEnhancer.EnhanceChart(chartRequested, deploySpec)

	return chartRequested, err
}

func (h helmDeployer) DeployApp(ctx context.Context, deploySpec DeploymentParams) error {

	// NewClient() defaults to using ~/.docker/config.json, but explictly state config path here for clarity
	// Hint: credStore defaults to file - if testing on dev machine ensure JSON isn't forcing credsStore (e.g. on osx oci lib will try to use osx keychain)
	regClient, _ := registry.NewClient(registry.ClientOptCredentialsFile(h.ociConfig.ConfigPath), registry.ClientOptDebug(true))

	// Initialize the action configuration
	helmActionConfiguration, err := h.getActionConfiguration(deploySpec.AppsNamespace, registryConfigOpt(regClient))
	if err != nil {
		return errors.WrapIf(err, "failed to initialize chart")
	}

	if h.isUpgrade(helmActionConfiguration, deploySpec.DeployID) {
		upgradeAction := action.NewUpgrade(&helmActionConfiguration)
		upgradeAction.Namespace = deploySpec.AppsNamespace

		// Setting Version is mandatory for OCI
		upgradeAction.Version = deploySpec.ApplicationVersion

		chartRequested, err := h.PullChart(&upgradeAction.ChartPathOptions, deploySpec)
		if err != nil {
			return errors.WrapIf(err, "failed to load chart")
		}

		chartValues, err := h.buildReleaseVars(deploySpec, chartRequested)
		if err != nil {
			return errors.Wrap(err, "failed to prepare configuration values")
		}

		release, err := upgradeAction.Run(deploySpec.DeployID, chartRequested, chartValues)
		if err != nil {
			return errors.WrapIf(err, "failed to install chart")
		}
		h.logger.Info("chart successfully upgraded", "release name", release.Name)
	} else {
		installAction := action.NewInstall(&helmActionConfiguration)
		installAction.ReleaseName = deploySpec.DeployID
		installAction.Namespace = deploySpec.AppsNamespace
		installAction.CreateNamespace = true

		// Setting Version is mandatory for OCI
		installAction.Version = deploySpec.ApplicationVersion

		chartRequested, err := h.PullChart(&installAction.ChartPathOptions, deploySpec)
		if err != nil {
			return errors.WrapIf(err, "failed to load chart")
		}

		chartValues, err := h.buildReleaseVars(deploySpec, chartRequested)
		h.logger.Info("chart values", chartValues)

		if err != nil {
			return errors.Wrap(err, "failed to prepare configuration values")
		}

		release, err := installAction.Run(chartRequested, chartValues)
		if err != nil {
			return errors.WrapIf(err, "failed to install chart")
		}
		h.logger.Info("chart successfully deployed", "release name", release.Name)
	}

	return nil
}

// buildReleaseVars constructs GBear-specific payload for helm charts
func (h helmDeployer) buildReleaseVars(deploySpec DeploymentParams, chart *chart.Chart) (map[string]interface{}, error) {
	// TODO: there should be some generic way for ingress configuration
	// TODO: other than relying on every app chart provides its own ingress template
	chartValues := make(map[string]interface{})
	chartValues["siteId"] = deploySpec.SiteID

	chartValues["ingress"] = map[string]interface{}{
		"enabled":      deploySpec.IngressURL != "" || deploySpec.IstioGateway != "",
		"istioGateway": deploySpec.IstioGateway,
		"servicePath":  deploySpec.ServicePath,
	}

	// TODO keeping a copy of config vars here for a while
	chartValues["config"] = deploySpec.Config

	chartValues["global"] = map[string]interface{}{
		"config": deploySpec.Config,
	}

	chartValues["networkConfig"] = h.networkConfig.Map()

	return chartValues, err
}

// UnDeployApp removes the application specified in the deployspec
func (h helmDeployer) UnDeployApp(ctx context.Context, deploySpec DeploymentParams) error {
	helmActionConfiguration, err := h.getActionConfiguration(deploySpec.AppsNamespace)
	if err != nil {
		return errors.WrapIf(err, "failed to uninstall chart")
	}

	_, uninstallErr := action.NewUninstall(&helmActionConfiguration).Run(deploySpec.DeployID)
	if uninstallErr != nil {
		if _, err := helmActionConfiguration.Releases.History(deploySpec.DeployID); err != nil {
			h.logger.Info("unable to uninstall, release not found", "name", deploySpec.DeployID)
			return nil
		}
		return errors.WrapIfWithDetails(uninstallErr, "failed to uninstall chart", "app", deploySpec.AppName)
	}

	return nil
}

// getActionConfiguration assembles an "in-cluster" action configuration to be used for helm operations
func (h helmDeployer) getActionConfiguration(namespace string, options ...ConfigOpt) (action.Configuration, error) {
	config := action.Configuration{}
	cfgFlags := genericclioptions.NewConfigFlags(true)
	cfgFlags.Namespace = &namespace
	err := config.Init(cfgFlags, namespace, "secret", func(format string, v ...interface{}) {})
	if err != nil {
		return action.Configuration{}, errors.WrapIf(err, "failed to initialize the action configuration")
	}

	for _, o := range options {
		o(&config)
	}

	return config, nil
}
