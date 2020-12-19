package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// ExistsPath is path exist
func ExistsPath(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// IsDirectory is path a directory
func IsDirectory(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}
	return false
}

// IsFile is path a file
func IsFile(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.Mode().IsRegular() {
		return true
	}
	return false
}

// ReadFile reads content of a file
func ReadFile(path string) (string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ReadJsonFile reads json content of file, unmarshal to an interface{}
func ReadJsonFile(path string, v interface{}) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}
