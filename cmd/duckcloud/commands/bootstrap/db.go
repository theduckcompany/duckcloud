package bootstrap

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func bootstrapDB(cmd *cobra.Command, folderPath string) *sql.DB {
	dbPath := path.Join(folderPath, dbFileName)

	_, err := os.OpenFile(dbPath, os.O_CREATE, 0o640)
	if err != nil {
		printErrAndExit(cmd, fmt.Errorf("failed to create %q: %w", dbPath, err))
	}

	db, err := storage.NewSQliteClient(&storage.Config{Path: dbPath})
	if err != nil {
		printErrAndExit(cmd, err)
	}

	err = storage.RunMigrations(db, nil)
	if err != nil {
		printErrAndExit(cmd, fmt.Errorf("migration error: %w", err))
	}

	cmd.Printf("Database initialized\n")

	return db
}
