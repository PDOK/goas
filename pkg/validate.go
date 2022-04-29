package pkg

import (
	"fmt"
	"strings"
)

func ValidateOGCStyles(ogcStyles *OGCStyles) error {
	var errors []string
	err := validateUniqueStyles(ogcStyles)
	if err != nil {
		errors = append(errors, err.Error())
	}
	err = validateDefaultStyle(ogcStyles)
	if err != nil {
		errors = append(errors, err.Error())
	}
	for _, metadata := range ogcStyles.StylesMetadata {
		err = validateStyleEncoding(metadata)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	if errors != nil {
		return fmt.Errorf("validation errors found: %s", strings.Join(errors, "; "))
	}
	return nil
}

// validateUniqueStyles Requirement 3D: The id member of each style SHALL be unique.
func validateUniqueStyles(ogcStyles *OGCStyles) error {
	var duplicateIds []string
	styleSet := make(map[string]bool)
	for _, metadata := range ogcStyles.StylesMetadata {
		_, ok := styleSet[metadata.Id]
		if !ok {
			styleSet[metadata.Id] = true
		} else {
			duplicateIds = append(duplicateIds, metadata.Id)
		}
	}
	if duplicateIds != nil {
		return fmt.Errorf("requirement 3D fails; found styles with duplicate ids: %s", strings.Join(duplicateIds, ", "))
	}
	return nil
}

// validateStyleEncoding Requirement 3E: Each style SHALL have at least one link to a style encoding supported for the style (link relation type: stylesheet) with the type attribute stating the media type of the style encoding.
func validateStyleEncoding(metadata StyleMetadata) error {
	for _, style := range metadata.Stylesheets {
		if style.Link.Rel == StylesheetRelation && style.Link.Type != nil {
			return nil
		}
	}
	return fmt.Errorf("requirement 3E fails; style %s stylesheet definition incorrect", metadata.Id)
}

// validateDefaultStyle Requirement 3G: The default member SHALL, if provided, be the id of one of the styles in the styles array.
func validateDefaultStyle(ogcStyles *OGCStyles) error {
	for _, metadata := range ogcStyles.StylesMetadata {
		if metadata.Id == ogcStyles.Default {
			return nil
		}
	}
	return fmt.Errorf("requirement 3G fails; default  %s not found in styles", ogcStyles.Default)
}

// TODO possible validation todos?:
// Requirement 4B The content of that response SHALL conform to the media type stated in the Content-Type header.

// Recommendation 2A:
//Sample data that can be used to illustrate the style SHOULD be represented as links with the following link relation types:
//enclosure for links to sample data that may be downloaded (e.g. a GeoPackage);
//collection for links to a Collection resource according to OGC API Common (e.g. /collections/{collectionId}; the collection may be available as features (tiled or not) or as gridded data);
//start for links to a Features resource according to OGC API Features (e.g. /collections/{collectionId}/items; the response may contain a next link to additional features);
//http://www.opengis.net/def/rel/ogc/1.0/tilesets-vector for a link to a Tile Sets resource (e.g. /collections/{collectionId}/tiles) with vector tiles.
//http://www.opengis.net/def/rel/ogc/1.0/tilesets-coverage for a link to a Tile Sets resource (e.g. /collections/{collectionId}/tiles) with coverage tiles.

// Recommendation 3A: If a style can be used to style multiple geospatial datasets that implement a common schema and where a canonical URI exists for the schema, a link with the link relation type http://www.opengis.net/def/rel/ogc/1.0/schema SHOULD be provided.
