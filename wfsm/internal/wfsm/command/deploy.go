package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/cisco-eti/wfsm/internal/deployer"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/cisco-eti/wfsm/internal/wfsm/manifest"
)

var deployLongHelp = `
This command takes two required flags: --manifestPath path/to/acpManifest
                                       --envFilePath path/to/envFile
An optional flag --deleteBuildFolders can be set to true or false to determine if the build folders should be deleted after deployment.
		
Examples:
- Deploy an agent with a manifest and environment file:
	wfsm deploy --manifestPath path/to/acpManifest --envFilePath path/to/envFile
`

const deployFail = "Deploy Status: Failed - %s"
const deployError string = "get failed"

const manifestPathFlag string = "manifestPath"
const envFilePathFlag string = "envFilePath"
const deleteBuildFoldersFlag string = "true"

// deployCmd represents the image build and run docker commands
var deployCmd = &cobra.Command{
	Use:   "deploy --manifestPath path/to/acpManifest --envFilePath path/to/envFile",
	Short: "Deploy an ACP agent",
	Long:  deployLongHelp,
	RunE: func(cmd *cobra.Command, args []string) error {

		manifestPath, _ := cmd.Flags().GetString(manifestPathFlag)
		envFilePath, _ := cmd.Flags().GetString(envFilePathFlag)
		deleteBuildFolders, _ := cmd.Flags().GetBool(deleteBuildFoldersFlag)

		err := runDeploy(manifestPath, envFilePath, deleteBuildFolders)
		if err != nil {
			util.OutputMessage(deployFail, err.Error())
			return fmt.Errorf(cmdErrorHelp, deployError)
		}
		return nil
	},
}

func init() {
	deployCmd.Flags().StringP(manifestPathFlag, "m", "", "Manifest file for the application")
	deployCmd.Flags().StringP(envFilePathFlag, "e", "", "Environment file for the application")

	deployCmd.Flags().BoolP(deleteBuildFoldersFlag, "d", true, "Delete build folders after deployment")

	deployCmd.MarkPersistentFlagRequired(envFilePathFlag)
	deployCmd.MarkPersistentFlagRequired(manifestPathFlag)

}

func runDeploy(manifestPath string, envFilePath string, deleteBuildFolders bool) error {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &logger

	ctx := logger.WithContext(context.Background())

	manifestSvc := manifest.NewManifestService(manifestPath)
	if err := manifestSvc.Validate(ctx); err != nil {
		return errors.New(fmt.Sprintf("invalid manifest: %v", err))
	}

	manifest, err := manifestSvc.GetManifest(ctx)
	if err != nil {
		return err
	}
	deployment := manifest.Deployment
	if deployment == nil {
		return errors.New("invalid agent manifest: no deployment found in manifest")
	}
	if len(deployment.DeploymentOptions) == 0 {
		return errors.New("invalid agent manifest: no deployment option found in manifest")
	}
	if len(deployment.DeploymentOptions) > 1 {
		return errors.New("invalid agent manifest: to many deployment options found in manifest")
	}

	//TODO get this from a command line options or env variable
	hostStorageFolder := os.Getenv("HOST_STORAGE_FOLDER")
	if hostStorageFolder == "" {
		homeDir, err := util.GetHomeDir()
		if err != nil {
			return errors.New("failed to get home directory")
		}

		hostStorageFolder = path.Join(homeDir, ".wfsm")
		// make sure the folder exists
		if _, err := os.Stat(hostStorageFolder); os.IsNotExist(err) {
			if err := os.Mkdir(hostStorageFolder, 0755); err != nil {
				return errors.New("failed to create host storage folder")
			}
		}
	}

	deployer := deployer.GetAgentDeployer(
		manifest, deployment.DeploymentOptions[0], envFilePath,
		hostStorageFolder, deleteBuildFolders)
	err = deployer.Deploy(ctx)
	if err != nil {
		return err
	}
	return nil
}
