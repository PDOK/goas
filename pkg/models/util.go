package models

import "fmt"

type StringSlice interface {
	ToString() []string
}

func unmarshalYaml(unmarshal func(interface{}) error, stringSlice StringSlice) (string, error) {
	var value string
	err := unmarshal(&value)
	if err != nil {
		return "", err
	}
	for _, item := range stringSlice.ToString() {
		if item == value {
			return value, nil
		}
	}

	return "", fmt.Errorf("could not unmarshal %s", value)
}
