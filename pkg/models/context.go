package models

import (
	"bytes"
)

type OGCStyles struct {
	BaseResource      string               `yaml:"base-resource"`
	Default           string               `yaml:"default,omitempty"`
	AdditionalFormats map[MediaType]Format `yaml:"additional-formats,omitempty"`
	StylesMetadata    []StyleMetadata      `yaml:"styles"`
}

type Document struct {
	Path      string
	MediaType MediaType
	Content   *bytes.Buffer
	Error     error
}

type Documents chan *Document

func (documents Documents) Add(document *Document, err error) bool {
	if err != nil {
		documents <- &Document{Error: err}
		return false
	}
	if document != nil {
		documents <- document
	}
	return true
}
