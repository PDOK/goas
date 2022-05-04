package models

import (
	"fmt"
	"log"
)

// http://www.opengis.net/def/rel/ogc/1.0/styles: Refers to a collection of styles.
type Styles struct {
	Default string  `json:"default,omitempty"`
	Styles  []Style `json:"styles"`
}

type Style struct {
	Id    string `yaml:"id" json:"id"`
	Title string `yaml:"title" json:"title,omitempty"`
	Links []Link `yaml:"links" json:"links"` // minimally: style encoding ("rel": "stylesheet", "type": "application/vnd.mapbox.style+json" || "type": "application/vnd.ogc.sld+xml;version=1.0"), style metadata ("rel": "describedby"), optionally: thumbnail (rel": "preview", "type": "image/png")
}

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

type StyleSheet struct {
	Title         *string `yaml:"title" json:"title,omitempty"`
	Version       *string `yaml:"version" json:"version,omitempty"`
	Specification *string `yaml:"specification" json:"specification,omitempty"`
	Native        *bool   `yaml:"native" json:"native,omitempty"`
	Link          Link    `yaml:"link" json:"link"`
}

type Link struct {
	AssetFilename *string      `yaml:"asset-filename" json:"-"`
	Href          *string      `yaml:"href" json:"href"`
	Rel           LinkRelation `yaml:"rel" json:"rel,omitempty"` // This is allowed to be empty according to the spec, but we leverage this
	Type          *MediaType   `yaml:"type" json:"type,omitempty"`
	Title         *string      `yaml:"title" json:"title,omitempty"`
	Hreflang      *string      `yaml:"hreflang" json:"hreflang,omitempty"`
	Length        *int         `yaml:"length" json:"length,omitempty"`
}

func (link Link) WithOtherRelation(otherRelation LinkRelation) *Link {
	link.Rel = otherRelation
	return &link
}

func (link Link) ToPath(identifier string, additionalFormats map[MediaType]Format) (string, error) {
	path, err := link.Rel.ToPath(identifier)
	if err != nil {
		return "", err
	}
	extension := link.Type.ToFormat(additionalFormats, false)
	if extension != "" {
		path = fmt.Sprintf("%s.%s", path, extension)
	}
	return path, nil
}

func (link *Link) UpdateHref(baseResource string, styleId string, additionalFormats map[MediaType]Format) error {
	url, err := link.Rel.ToUrl(baseResource, styleId)
	if err != nil {
		return err
	}
	format := link.Type.ToFormat(additionalFormats, true)
	if format != "" {
		urlWithFormatExtension := fmt.Sprintf("%s.%s", *url, format)
		url = &urlWithFormatExtension
	}
	if link.Href != nil {
		log.Printf("link href `%s` not empty, overwriting with: `%s`", *link.Href, *url)
	}
	link.Href = url
	return nil
}

type PropertiesSchema struct{} // TODO implement later
