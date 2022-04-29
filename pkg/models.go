package pkg

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
)

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

type LinkRelation string
type LinkRelations []LinkRelation

// LinkRelation All known relations, not all are (as yet) used.
//  Taken from: OGC API styles - 5.2
const (
	AlternateRelation       LinkRelation = "alternate"                                               // Refers to a substitute for this context.
	CollectionRelation      LinkRelation = "collection"                                              // The target IRI points to a resource which represents the collection resource for the context IRI.
	DescribedbyRelation     LinkRelation = "describedby"                                             // Metadata = Refers to a resource providing information about the link’s context.
	EnclosureRelation       LinkRelation = "enclosure"                                               // Sample data = Identifies a related resource that is potentially large and might require special handling.
	PreviewRelation         LinkRelation = "preview"                                                 // Thumbnail = Refers to a resource that provides a preview of the link’s context.
	SelfRelation            LinkRelation = "self"                                                    // Conveys an identifier for the link’s context.
	ServiceDescRelation     LinkRelation = "service-desc"                                            // Identifies service description for the context that is primarily intended for consumption by machines.
	ServiceDocRelation      LinkRelation = "service-doc"                                             // Identifies service documentation for the context that is primarily intended for human consumption.
	StartRelation           LinkRelation = "start"                                                   // OGC API Features = Refers to the first resource in a collection of resources.
	StylesheetRelation      LinkRelation = "stylesheet"                                              // Refers to a stylesheet.
	SchemaRelation          LinkRelation = "http://www.opengis.net/def/rel/ogc/1.0/schema"           // Refers to a schema that data has to conform to to be suitable for use with the link’s context. - (OGC API Styles - Recommendation 3A)
	StylesRelation          LinkRelation = "http://www.opengis.net/def/rel/ogc/1.0/styles"           // Refers to a collection of styles.
	ConformanceRelation     LinkRelation = "http://www.opengis.net/def/rel/ogc/1.0/conformance"      // Refers to resource that identifies the specifications that the link’s context conforms to.
	TilesetsVectorRelation  LinkRelation = "http://www.opengis.net/def/rel/ogc/1.0/tilesets-vector"  // The target IRI points to a resource that describes how to provide tile sets of the context resource in vector format.
	TilesetCoverageRelation LinkRelation = "http://www.opengis.net/def/rel/ogc/1.0/tileset-coverage" // The target IRI points to a resource that describes how to provide tile sets of the context resource in coverage format.
)

var linkRelations = LinkRelations{
	AlternateRelation, CollectionRelation, DescribedbyRelation, EnclosureRelation, PreviewRelation, SelfRelation,
	ServiceDescRelation, ServiceDocRelation, StartRelation, StylesheetRelation, SchemaRelation, StylesRelation,
	ConformanceRelation, TilesetsVectorRelation, TilesetCoverageRelation,
}

func (linkRelations LinkRelations) ToString() (result []string) {
	for _, linkRelation := range linkRelations {
		result = append(result, string(linkRelation))
	}
	return result
}

func (linkRelation LinkRelation) ToPath(identifier string) (string, error) {
	switch linkRelation {
	case StylesheetRelation:
		return fmt.Sprintf(styleResource, identifier), nil
	case DescribedbyRelation:
		return fmt.Sprintf(styleMetadataResource, identifier), nil
	case PreviewRelation:
		return fmt.Sprintf(stylesPreviewResource, identifier), nil
	default:
		return "", fmt.Errorf("no path known for link relation: %s", linkRelation)
	}
}

func (linkRelation LinkRelation) MustToPath(identifier string) string {
	path, err := linkRelation.ToPath(identifier)
	if err != nil {
		panic(err)
	}
	return path
}

func (linkRelation LinkRelation) ToUrl(baseResource string, identifier string) (*string, error) {
	path, err := linkRelation.ToPath(identifier)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/%s", baseResource, path)
	return &url, nil
}

func (linkRelation LinkRelation) MustToUrl(baseResource string, identifier string) *string {
	url, err := linkRelation.ToUrl(baseResource, identifier)
	if err != nil {
		panic(err)
	}
	return url
}

// UnmarshalYAML unmarshals a yaml string to the LinkRelation value
func (linkRelation *LinkRelation) UnmarshalYAML(unmarshal func(interface{}) error) error {
	result, err := unmarshalYaml(unmarshal, linkRelations)
	if err != nil {
		return fmt.Errorf("unknown link relation with error: %s", err.Error())
	}
	*linkRelation = LinkRelation(result)
	return nil
}

type MediaType string
type Format string

const (
	JsonMediaType   MediaType = "application/json"
	HtmlMediaType   MediaType = "text/html"
	SldMediaType    MediaType = "application/vnd.ogc.sld+xml"
	MapboxMediaType MediaType = "application/vnd.mapbox.style+json"

	JsonFormat   Format = "json"
	HtmlFormat   Format = "html"
	SldFormat    Format = "sld"
	MapboxFormat Format = "mapbox"

	mediaTypeSeperator     = ";"
	mediaTypePartSeperator = "="

	queryParams = "?"
	formatQuery = "f=%s"
)

var knownFormats = map[MediaType]Format{
	JsonMediaType:   JsonFormat,
	HtmlMediaType:   HtmlFormat,
	SldMediaType:    SldFormat,
	MapboxMediaType: MapboxFormat,
}

var versionRegex = regexp.MustCompile(`\d+`)

func (m MediaType) SplitParams() (MediaType, map[string]string) {
	mediatypeParts := strings.Split(string(m), mediaTypeSeperator)
	params := make(map[string]string)
	root := MediaType(mediatypeParts[0])
	if len(mediatypeParts) > 1 {
		for _, mediatypePart := range mediatypeParts[1:] {
			paramParts := strings.Split(mediatypePart, mediaTypePartSeperator)
			if len(paramParts) == 1 {
				params[paramParts[0]] = ""
			} else if len(paramParts) == 2 {
				params[paramParts[0]] = paramParts[1]
			} else {
				log.Printf("mediatype %s has unknown params", m)
			}
		}
	}
	return root, params
}

func (m MediaType) ToFormat(additionalFormats map[MediaType]Format, versioned bool) Format {
	root, params := m.SplitParams()
	format, ok := knownFormats[root]
	if !ok {
		if additionalFormats == nil {
			return ""
		}
		format, ok = additionalFormats[root]
		if !ok {
			return ""
		}
	}
	if versioned {
		version, ok := params["version"]
		if ok {
			versionDigits := strings.Join(versionRegex.FindAllString(version, -1), "")
			format = Format(fmt.Sprintf("%s%s", format, versionDigits))
		}
	}
	return format
}

func (format Format) ToQuery() string {
	if format != "" {
		return fmt.Sprintf(formatQuery, format)
	}
	return ""
}

type GeometryType string
type GeometryTypes []GeometryType

const (
	Points   GeometryType = "points"
	Lines    GeometryType = "lines"
	Polygons GeometryType = "polygons"
	Solids   GeometryType = "solids"
	Any      GeometryType = "any"
)

var geometryTypes = GeometryTypes{Points, Lines, Polygons, Solids, Any}

func (geometryTypes GeometryTypes) ToString() (result []string) {
	for _, geometryType := range geometryTypes {
		result = append(result, string(geometryType))
	}
	return result
}

func (geometryType *GeometryType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	result, err := unmarshalYaml(unmarshal, geometryTypes)
	if err != nil {
		return fmt.Errorf("unknown geometry type with error: %s", err.Error())
	}
	*geometryType = GeometryType(result)
	return nil
}

type DataType string
type DataTypes []DataType

const (
	Vector   DataType = "vector"
	Map      DataType = "map"
	Coverage DataType = "coverage"
)

var dataTypes = DataTypes{Vector, Map, Coverage}

func (dataTypes DataTypes) ToString() (result []string) {
	for _, dataType := range dataTypes {
		result = append(result, string(dataType))
	}
	return result
}

func (dataType *DataType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	result, err := unmarshalYaml(unmarshal, dataTypes)
	if err != nil {
		return fmt.Errorf("unknown geometry type %s", result)
	}
	*dataType = DataType(result)
	return nil
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

func (link Link) WithOtherRelation(otherRelation LinkRelation) Link {
	link.Rel = otherRelation
	return link
}

func (link *Link) UpdateHref(baseResource string, styleId string, additionalFormats map[MediaType]Format) error {
	url, err := link.Rel.ToUrl(baseResource, styleId)
	if err != nil {
		return err
	}
	formatParams := link.Type.ToFormat(additionalFormats, true).ToQuery()
	if formatParams != "" {
		urlWithFormatquery := fmt.Sprintf("%s%s%s", *url, queryParams, formatParams)
		url = &urlWithFormatquery
	}
	if link.Href != nil {
		log.Printf("link href `%s` not empty, overwriting with: `%s`", *link.Href, *url)
	}
	link.Href = url
	return nil
}

type OGCStyles struct {
	BaseResource      string               `yaml:"base-resource"`
	Default           string               `yaml:"default,omitempty"`
	AdditionalFormats map[MediaType]Format `yaml:"additional-formats,omitempty"`
	StylesMetadata    []StyleMetadata      `yaml:"styles"`
}

type Style struct {
	Id    string `yaml:"id" json:"id"`
	Title string `yaml:"title" json:"title,omitempty"`
	Links []Link `yaml:"links" json:"links"` // minimally: style encoding ("rel": "stylesheet", "type": "application/vnd.mapbox.style+json" || "type": "application/vnd.ogc.sld+xml;version=1.0"), style metadata ("rel": "describedby"), optionally: thumbnail (rel": "preview", "type": "image/png")
}

// http://www.opengis.net/def/rel/ogc/1.0/styles: Refers to a collection of styles.
type Styles struct {
	Default string  `json:"default,omitempty"`
	Styles  []Style `json:"styles"`
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

type PropertiesSchema struct{} // TODO implement later

type Document struct {
	Path      string
	MediaType MediaType
	Content   *bytes.Buffer
	Error     error
}

type Documents chan *Document

func (documents Documents) HandleError(err error) {
	documents <- &Document{Error: err}
}
