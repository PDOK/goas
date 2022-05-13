package models

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

const (
	StylesResource        = "styles"
	StyleResource         = "styles/%s"
	StyleMetadataResource = "styles/%s/metadata"
	ResourceResource      = "resources/%s" // this is not clearly specified in the OGC API Styles spec, taken from the examples
)

type LinkRelation string
type LinkRelations []LinkRelation

// LinkRelation All known relations, not all are (as yet) used.
//  Taken from: OGC API styles - 5.2
const (
	AlternateRelation       LinkRelation = "alternate"                                               // Refers to a substitute for this context. (Refers to a substitute for this context [IANA]. Refers to a representation of the current resource which is encoded using another media type (the media type is specified in the type link attribute).)
	CollectionRelation      LinkRelation = "collection"                                              // The target IRI points to a resource which represents the collection resource for the context IRI.
	DescribedbyRelation     LinkRelation = "describedby"                                             // Metadata = Refers to a resource providing information about the link’s context.
	EnclosureRelation       LinkRelation = "enclosure"                                               // Sample data = Identifies a related resource that is potentially large and might require special handling.
	PreviewRelation         LinkRelation = "preview"                                                 // Thumbnail = Refers to a resource that provides a preview of the link’s context.
	SelfRelation            LinkRelation = "self"                                                    // Conveys an identifier for the link’s context.
	ServiceDescRelation     LinkRelation = "service-desc"                                            // Identifies service description for the context that is primarily intended for consumption by machines.
	ServiceDocRelation      LinkRelation = "service-doc"                                             // Identifies service documentation for the context that is primarily intended for human consumption.
	StartRelation           LinkRelation = "start"                                                   // OGC API Features = Refers to the first resource in a collection of resources.
	StylesheetRelation      LinkRelation = "stylesheet"                                              // Refers to a stylesheet.
	PreloadRelation         LinkRelation = "preload"                                                 // Refers to a resource that should be loaded early in the processing of the link's context, without blocking rendering. E.g. Fonts or sprites
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
	case StylesRelation:
		return StylesResource, nil
	case StylesheetRelation:
		return fmt.Sprintf(StyleResource, identifier), nil
	case DescribedbyRelation:
		return fmt.Sprintf(StyleMetadataResource, identifier), nil
	case PreviewRelation, PreloadRelation:
		return fmt.Sprintf(ResourceResource, identifier), nil
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

const (
	JsonMediaType   MediaType = "application/json"
	HtmlMediaType   MediaType = "text/html"
	SldMediaType    MediaType = "application/vnd.ogc.sld+xml"
	MapboxMediaType MediaType = "application/vnd.mapbox.style+json"
	PngMediaType    MediaType = "image/png"

	mediaTypeSeperator     = ";"
	mediaTypePartSeperator = "="

	formatQuery = "f=%s"
)

var (
	JsonFormat       = Format{JsonMediaType, "json", "json"}
	HtmlFormat       = Format{HtmlMediaType, "html", "html"}
	SldFormat        = Format{SldMediaType, "sld", "sld"}
	MapboxFormat     = Format{MapboxMediaType, "mapbox", "mapbox.json"}
	PngFormat        = Format{PngMediaType, "png", "png"}
	knownBaseFormats = []Format{JsonFormat, HtmlFormat, SldFormat, MapboxFormat, PngFormat}
)

func GetFormat(format string) (Format, bool) {
	for _, knownFormat := range knownBaseFormats {
		if knownFormat.Name == format {
			return knownFormat, true
		}
	}
	return Format{}, false
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

func (m MediaType) ToFormat(additionalFormats []Format, versioned bool) Format {
	baseFormat := Format{}
	root, params := m.SplitParams()
	for _, format := range append(knownBaseFormats, additionalFormats...) {
		if root == format.MediaType {
			baseFormat = format
			break
		}
	}

	if versioned {
		version, ok := params["version"]
		if ok {
			versionDigits := strings.Join(versionRegex.FindAllString(version, -1), "")
			baseFormat.Name = fmt.Sprintf("%s%s", baseFormat.Name, versionDigits)
		}
	}
	return baseFormat
}

func (f Format) ToQuery() string {
	if f.Name == "" {
		return ""
	}
	return fmt.Sprintf(formatQuery, f.Name)
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
