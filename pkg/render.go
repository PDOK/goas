package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Renderer func(obj interface{}, path string) (*Document, error)

func Render(obj interface{}, path string, format Format) (*Document, error) {
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

func getRenderer(format Format) (Renderer, error) {
	switch format {
	case JsonFormat:
		return jsonRenderer, nil
	default:
		return nil, fmt.Errorf("format: %v not implemented", format)
	}
}

func jsonRenderer(obj interface{}, path string) (*Document, error) {
	content := new(bytes.Buffer)
	enc := json.NewEncoder(content)
	enc.SetEscapeHTML(false)
	err := enc.Encode(obj)
	if err != nil {
		return nil, fmt.Errorf("error: %v, could not render document to file", err)
	}
	return &Document{path, JsonMediaType, content, nil}, nil
}
