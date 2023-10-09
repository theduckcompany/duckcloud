package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/internal/service/config"
)

const (
	folderName   = "duckcloud"
	dbFileName   = "db.sqlite"
	filesDirName = "files"
)

func bootstrapFolder(cmd *cobra.Command, dir string) string {
	var err error
	var folderPath string

	if dir != "" {
		folderPath, err = filepath.Abs(dir)
		if err != nil {
			cmd.PrintErrln(fmt.Sprintf(`invalid path %q: %s`, folderPath, err))
			os.Exit(1)
		}
	}

	if folderPath == "" {
		folderPath, err = xdg.SearchDataFile(folderName)
	}

	if folderPath == "" {
		folderPath = path.Join(xdg.DataHome, folderName)
	}

	user, err := user.Current()
	if err != nil {
		printErrAndExit(cmd, err)
	}

	var confirm bool
	err = survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("The server folder will be created at %q with the user %q", folderPath, user.Name),
		Default: true,
	}, &confirm)
	if err != nil {
		printErrAndExit(cmd, err)
	}

	if !confirm {
		printErrAndExit(cmd, errors.New("aborted"))
	}

	err = os.MkdirAll(folderPath, 0o755)
	if err != nil {
		printErrAndExit(cmd, fmt.Errorf("failed to create the folder: %w", err))
	}

	fmt.Printf("Folder created at %s\n", folderPath)

	return folderPath
}

func setupFolderPath(cmd *cobra.Command, configSvc config.Service, folderPath string) {
	folder, err := configSvc.Get(cmd.Context(), config.FSDataFolder)
	if err != nil {
		printErrAndExit(cmd, err)
	}

	if folder != "" {
		fmt.Println("Folder localisation saved")
		return
	}

	err = configSvc.SetDataFolder(cmd.Context(), folderPath)
	if err != nil {
		printErrAndExit(cmd, fmt.Errorf("failed to save the data folder path: %w", err))
	}
}
