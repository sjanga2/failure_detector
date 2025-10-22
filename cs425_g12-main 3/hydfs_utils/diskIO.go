package hydfs_utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var HyDFSLogger *log.Logger

// logger file
func InitializeLogger(machineName string) {
	logPath := filepath.Join("/home/shared/hydfs/logs", machineName+".log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("could not open log file: %v", err)
	}
	HyDFSLogger = log.New(file, "", log.Ldate|log.Ltime|log.Lmicroseconds)
}

// init the Hydfs directories
func InitHyDFSDir() error {
	dirs := []string{"/home/shared/hydfs", "/home/shared/hydfs/data", "/home/shared/hydfs/logs"}

	for _, d := range dirs {
		// permissions set to rwxr-xr-x
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

func WriteLocalFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	// make sure the directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories for %s: %v", path, err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write data to file: %v", err)
	}
	return nil
}

func ReadLocalFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	return data, nil
}
