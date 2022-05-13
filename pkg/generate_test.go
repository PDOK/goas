package pkg

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pdok/goas/pkg/models"
	"github.com/stretchr/testify/require"
)

var whiteSpace = []string{"\n", "\t", " "}

func bytesToComparableString(content *bytes.Buffer) string {
	result := content.String()
	for _, space := range whiteSpace {
		result = strings.Replace(result, space, "", -1)
	}
	return result
}

func TestParseConfig(t *testing.T) {
	cfg, err := ParseConfig("../examples/config.yaml")
	require.Nil(t, err)
	require.Len(t, cfg.StylesMetadata, 1)
	require.Len(t, cfg.StylesMetadata[0].Stylesheets, 3)
	require.Len(t, cfg.StylesMetadata[0].Links, 1)
}

func TestGenerateDocuments(t *testing.T) {
	config, _ := ParseConfig("../examples/config.yaml")
	documents := GenerateDocuments(config, "../examples/assets", []models.Format{models.JsonFormat})
	expectedDocuments := []models.Document{
		{
			"resources/thumbnail.png",
			"image/png",
			bytes.NewBuffer([]byte("")),
			nil},
		{
			"styles/night.mapbox.json",
			"application/vnd.mapbox.style+json",
			bytes.NewBuffer([]byte( //language=json
				`{"MAPBOX_STYLE": "https://example.org/catalog/1.0"}`)),
			nil},
		{
			"styles/night.sld",
			"application/vnd.ogc.sld+xml;version=1.0",
			bytes.NewBuffer([]byte( //language=xml
				`<root href="https://example.org/catalog/1.0">SLD</root>`)),
			nil},
		{
			"styles/night.custom.json",
			"application/vnd.custom.style+json",
			bytes.NewBuffer([]byte( //language=text
				`Custom Style = https://example.org/catalog/1.0`)),
			nil},
		{
			"styles/night/metadata.json",
			"application/json",
			bytes.NewBuffer([]byte( //language=json
				`{
				  "id": "night",
				  "title": "Topographic night style",
				  "description": "This topographic basemap style is designed to be used in situations with low ambient light. The style supports datasets based on the TDS 6.1 specification.",
				  "keywords": [
					"basemap",
					"TDS",
					"TDS 6.1",
					"OGC API"
				  ],
				  "pointOfContact": "John Doe",
				  "license": "MIT",
				  "created": "2019-01-01T10:05:00Z",
				  "updated": "2019-01-01T11:05:00Z",
				  "scope": "style",
				  "version": "1.0.0",
				  "stylesheets": [
					{
					  "title": "Mapbox Style",
					  "version": "8",
					  "specification": "https://docs.mapbox.com/mapbox-gl-js/style-spec/",
					  "native": true,
					  "link": {
						"href": "https://example.org/catalog/1.0/styles/night?f=mapbox",
						"rel": "stylesheet",
						"type": "application/vnd.mapbox.style+json"
					  }
					},
					{
					  "title": "OGC SLD",
					  "version": "1.0",
					  "native": false,
					  "link": {
						"href": "https://example.org/catalog/1.0/styles/night?f=sld10",
						"rel": "stylesheet",
						"type": "application/vnd.ogc.sld+xml;version=1.0"
					  }
					},
					{
					  "title": "Custom Style",
					  "native": true,
					  "link": {
						"href": "https://example.org/catalog/1.0/styles/night?f=custom",
						"rel": "stylesheet",
						"type": "application/vnd.custom.style+json"
					  }
					}
				  ],
				  "layers": [
					{
					  "id": "VegetationSrf",
					  "geometryType": "polygons",
					  "sampleData": {
						"href": "https://demo.ldproxy.net/daraa/collections/VegetationSrf/items?f=json&limit=100",
						"rel": "start",
						"type": "application/geo+json"
					  }
					},
					{
					  "id": "hydrographycrv",
					  "geometryType": "lines",
					  "sampleData": {
						"href": "https://services.interactive-instruments.de/vtp/daraa/collections/hydrographycrv/items?f=json&limit=100",
						"rel": "start",
						"type": "application/geo+json"
					  }
					}
				  ],
				  "links": [
					{
					  "href": "https://example.org/catalog/1.0/resources/night.png",
					  "rel": "preview",
					  "type": "image/png",
					  "title": "thumbnail of the night style applied to OSM data from Daraa, Syria"
					},
					{
					  "href": "https://example.org/catalog/1.0/styles/night/metadata",
					  "rel": "self",
					  "title": "Style Metadata for night"
					}
				  ]
				}`),
			),
			nil},
		{
			"styles.json",
			"application/json",
			bytes.NewBuffer([]byte( //language=json
				`{
				  "default": "night",
				  "styles": [
					{
					  "id": "night",
					  "title": "Topographic night style",
					  "links": [
						{
						  "href": "https://example.org/catalog/1.0/resources/night.png",
						  "rel": "preview",
						  "type": "image/png",
						  "title": "thumbnail of the night style applied to OSM data from Daraa, Syria"
						},
						{
						  "href": "https://example.org/catalog/1.0/styles/night/metadata",
						  "rel": "describedby",
						  "title": "Style Metadata for night"
						},
						{
						  "href": "https://example.org/catalog/1.0/styles/night?f=mapbox",
						  "rel": "stylesheet",
						  "type": "application/vnd.mapbox.style+json"
						},
						{
						  "href": "https://example.org/catalog/1.0/styles/night?f=sld10",
						  "rel": "stylesheet",
						  "type": "application/vnd.ogc.sld+xml;version=1.0"
						},
						{	
						  "href": "https://example.org/catalog/1.0/styles/night?f=custom",
						  "rel": "stylesheet",
						  "type": "application/vnd.custom.style+json"
						}
					  ]
					}
				  ]
				}`)),
			nil},
	}
	for _, expectedDocument := range expectedDocuments {
		document := <-documents
		require.Equal(t, expectedDocument.Path, document.Path)
		require.Equal(t, expectedDocument.MediaType, document.MediaType)
		require.Equal(t, bytesToComparableString(expectedDocument.Content), bytesToComparableString(document.Content))
		require.Equal(t, expectedDocument.Error, document.Error)
	}
}

func TestGenerateDocumentsMinimalConfig(t *testing.T) {
	config, _ := ParseConfig("../examples/minimal_config.yaml")
	documents := GenerateDocuments(config, "../examples/assets", []models.Format{models.JsonFormat})
	expectedDocuments := []models.Document{
		{
			"styles/night/metadata.json",
			"application/json",
			bytes.NewBuffer([]byte( //language=json
				`{
					"id": "night",
					"title": "Topographic night style",
					"links": [
					  {
					    "href": "https://example.org/catalog/1.0/styles/night/metadata",
					    "rel": "self",
					    "title": "Style Metadata for night"
					  }
					]
				}`)),
			nil},
		{
			"styles.json",
			"application/json",
			bytes.NewBuffer([]byte( //language=json
				`{
 				  "styles": [
  					{
					  "id": "night",
					  "title": "Topographic night style",
					  "links": [
						{
						  "href": "https://example.org/catalog/1.0/styles/night/metadata",
						  "rel": "describedby",
						  "title": "Style Metadata for night"
						}
					  ]
					}
				  ]
				}`)),
			nil},
	}
	for _, expectedDocument := range expectedDocuments {
		document := <-documents
		require.Equal(t, expectedDocument.Path, document.Path)
		require.Equal(t, expectedDocument.MediaType, document.MediaType)
		require.Equal(t, bytesToComparableString(expectedDocument.Content), bytesToComparableString(document.Content))
		require.Equal(t, expectedDocument.Error, document.Error)
	}
}
