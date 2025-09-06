package winbackupchecker

import (
	"fmt"
	"os"
	"path/filepath"
)

func ScanBackupDir(root string) error {
    fmt.Println("Scanning backup root:", root)

    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Check for presence of BackupSpecs.xml (a key Windows Backup file)
        if filepath.Base(path) == "BackupSpecs.xml" {
            fmt.Println("Found backup set:", filepath.Dir(path))
        }

        return nil
    })

    if err != nil {
        return fmt.Errorf("scan failed: %w", err)
    }

    return nil
}
