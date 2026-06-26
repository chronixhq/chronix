package app

import (
	"os"
	"path/filepath"
)

func setWorkingDir() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	err = os.Chdir(exPath)
	if err != nil {
		panic(err)
	}
}
