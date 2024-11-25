package controllers

import (
	"path/filepath"
)

const UNKNOWN_FILE_TYPE = "bin"

func getFileType(fileName string) string {
	fileExtension := filepath.Ext(fileName)
	if len(fileExtension) > 0 {
		return fileExtension[1:]
	}

	return UNKNOWN_FILE_TYPE
}

// Get size in bytes and convert to gigabytes
func getFileSize(file []byte) float64 {
	return float64(len(file)) / (1024 * 1024 * 1024)
}
