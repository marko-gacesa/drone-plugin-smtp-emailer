package plugin

import "path/filepath"

func joinPath(basePath, fileName string) string {
	if filepath.IsAbs(fileName) {
		return fileName
	}

	return filepath.Join(basePath, fileName)
}
