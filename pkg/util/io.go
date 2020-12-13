package util

import (
	"encoding/json"
	"io/ioutil"
)

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
