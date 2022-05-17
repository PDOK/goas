package models

import (
	"fmt"
	"log"
	"strings"
)

// Styles based on OGC API Styles Requirement 3B -  http://www.opengis.net/def/rel/ogc/1.0/styles: Refers to a collection of styles.
type Styles struct {
	Default string  `json:"default,omitempty"`
	Styles  []Style `json:"styles"`
}

// Style based on OGC API Styles Requirement 3B
type Style struct {
	Id    string `yaml:"id" json:"id"`
	Title string `yaml:"title" json:"title,omitempty"`
	Links []Link `yaml:"links" json:"links"` // minimally: style encoding ("rel": "stylesheet", "type": "application/vnd.mapbox.style+json" || "type": "application/vnd.ogc.sld+xml;version=1.0"), style metadata ("rel": "describedby"), optionally: thumbnail (rel": "preview", "type": "image/png")
}

// StyleMetadata based on OGC API Styles Requirement 7B
type StyleMetadata struct {
	Id             string       `yaml:"id" json:"id"`
	Title          *string      `yaml:"title" json:"title,omitempty"`
	Description    *string      `yaml:"description" json:"description,omitempty"`
	Keywords       []string     `yaml:"keywords" json:"keywords,omitempty"`
	PointOfContact *string      `yaml:"point-of-contact" json:"pointOfContact,omitempty"`
	License        *string      `yaml:"license" json:"license,omitempty"`
	Created        *string      `yaml:"created" json:"created,omitempty"`
	Updated        *string      `yaml:"updated" json:"updated,omitempty"`
	Scope          *string      `yaml:"scope" json:"scope,omitempty"`
	Version        *string      `yaml:"version" json:"version,omitempty"`
	Stylesheets    []StyleSheet `yaml:"stylesheets" json:"stylesheets,omitempty"`
	Layers         []struct {
		Id           string        `yaml:"id" json:"id"`
		GeometryType *GeometryType `yaml:"type" json:"geometryType,omitempty"`
		SampleData   Link          `yaml:"sample-data" json:"sampleData,omitempty"`
		// TODO: the Properties schema is a stub and can be an implementation of: https://raw.githubusercontent.com/OAI/OpenAPI-Specification/master/schemas/v3.0/schema.json#/definitions/Schema
		PropertiesSchema *PropertiesSchema `yaml:"properties-schema" json:"propertiesSchema,omitempty"`
	} `yaml:"layers" json:"layers,omitempty"`
	Links []Link `yaml:"links" json:"links,omitempty"`
}

// StyleSheet based on OGC API Styles Requirement 7B
type StyleSheet struct {
	Title         *string `yaml:"title" json:"title,omitempty"`
	Version       *string `yaml:"version" json:"version,omitempty"`
	Specification *string `yaml:"specification" json:"specification,omitempty"`
	Native        *bool   `yaml:"native" json:"native,omitempty"`
	Link          Link    `yaml:"link" json:"link"`
}

// Link based on OGC API Features - http://schemas.opengis.net/ogcapi/features/part1/1.0/openapi/schemas/link.yaml - as referenced by OGC API Styles Requirements 3B and 7B
type Link struct {
	AssetFilename *string      `yaml:"asset-filename" json:"-"`
	Href          *string      `yaml:"href" json:"href"`
	Rel           LinkRelation `yaml:"rel" json:"rel,omitempty"` // This is allowed to be empty according to the spec, but we leverage this
	Type          *MediaType   `yaml:"type" json:"type,omitempty"`
	Title         *string      `yaml:"title" json:"title,omitempty"`
	Hreflang      *string      `yaml:"hreflang" json:"hreflang,omitempty"`
	Length        *int         `yaml:"length" json:"length,omitempty"`
}

type Format struct {
	MediaType MediaType `yaml:"media-type"`
	Name      string    `yaml:"name"`
	Extension string    `yaml:"extension"`
}

func (link Link) WithOtherRelation(otherRelation LinkRelation) *Link {
	link.Rel = otherRelation
	return &link
}

func (link Link) ToPath(identifier string, additionalFormats []Format) (string, error) {
	path, err := link.Rel.ToPath(identifier)
	if err != nil {
		return "", err
	}
	format := link.Type.ToFormat(additionalFormats, false)
	if format.Extension != "" && !strings.HasSuffix(path, format.Extension) {
		path = fmt.Sprintf("%s.%s", path, format.Extension)
	}
	return path, nil
}

func (link *Link) UpdateHref(baseResource string, styleId string, additionalFormats []Format, withQuery bool, withExtension bool) error {
	if withQuery && withExtension {
		return fmt.Errorf("href may not contain both a format query parameter and extension")
	}
	url, err := link.Rel.ToUrl(baseResource, styleId)
	if err != nil {
		return err
	}
	format := link.Type.ToFormat(additionalFormats, true)
	formatString := ""
	formatting := ""
	if withQuery {
		formatString = format.ToQuery()
		formatting = "%s?%s"
	} else if withExtension {
		formatString = format.Extension
		formatting = "%s.%s"
	}

	if formatString != "" {
		urlWithFormatquery := fmt.Sprintf(formatting, *url, formatString)
		url = &urlWithFormatquery
	}
	if link.Href != nil {
		log.Printf("link href `%s` not empty, overwriting with: `%s`", *link.Href, *url)
	}
	link.Href = url
	return nil
}

type PropertiesSchema struct{} // TODO implement later
