package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pdok/goas/pkg/models"
)

type Renderer func(obj interface{}, path string) (*models.Document, error)

func Render(obj interface{}, path string, format models.Format) (*models.Document, error) {
	renderer, err := getRenderer(format)
	if err != nil {
		return nil, err
	}
	document, err := renderer(obj, path)
	if err != nil {
		return nil, err
	}
	return document, nil
}

func getRenderer(format models.Format) (Renderer, error) {
	switch format {
	case models.JsonFormat:
		return jsonRenderer, nil
	default:
		return nil, fmt.Errorf("format: %v not implemented", format)
	}
}

func jsonRenderer(obj interface{}, path string) (*models.Document, error) {
	content := new(bytes.Buffer)
	enc := json.NewEncoder(content)
	enc.SetEscapeHTML(false)
	err := enc.Encode(obj)
	if err != nil {
		return nil, fmt.Errorf("error: %v, could not render document to file", err)
	}
	return &models.Document{Path: path, MediaType: models.JsonMediaType, Content: content, Error: nil}, nil
}
