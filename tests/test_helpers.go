// Package tests provides helper functions for testing
package tests

import (
	"archive/zip"
	"os"
)

// Helper function to create mock ZIP files for testing
func createMockZipFile(zipPath, csvFileName, csvContent string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	csvWriter, err := zipWriter.Create(csvFileName)
	if err != nil {
		return err
	}

	_, err = csvWriter.Write([]byte(csvContent))
	return err
}
