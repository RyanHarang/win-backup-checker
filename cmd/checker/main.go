package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
    backupRoot := `D:\Backups\WindowsImageBackup` // TODO: make configurable

    err := filepath.Walk(backupRoot, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() && filepath.Base(path) == "Catalog" {
            fmt.Println("Found backup catalog at:", path)
        }
        return nil
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Scan complete.")
}
