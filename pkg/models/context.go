package models

import (
	"bytes"
)

type StylesConfig struct {
	BaseResource      string            `yaml:"base-resource"`
	Default           string            `yaml:"default,omitempty"`
	AdditionalFormats []Format          `yaml:"additional-formats,omitempty"`
	AdditionalAssets  []AdditionalAsset `yaml:"additional-assets,omitempty"`
	StylesMetadata    []StyleMetadata   `yaml:"styles"`
}

type AdditionalAsset struct {
	Path      string    `yaml:"path"`
	MediaType MediaType `yaml:"media-type"`
}

type Document struct {
	Path      string
	MediaType MediaType
	Content   *bytes.Buffer
}
