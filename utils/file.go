package utils

import (
	"os"
	"path/filepath"
)

func CreateFileIfNotExists(filePath, exampleConfig string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		if err := os.WriteFile(filePath, []byte(exampleConfig), 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// FileExists 判断所给路径文件/文件夹是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func MkdirIfNotExists(path string) error {
	if !FileExists(path) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
