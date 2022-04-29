package pkg

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
	"text/template"
)

const (
	stylesResource        = "styles"
	styleResource         = "styles/%s"
	styleMetadataResource = "styles/%s/metadata"
	stylesPreviewResource = "resources/%s" // this is not clearly specified in the OGC API Styles spec, taken from the examples

	documentChanSize = 10
)

func ParseConfig(configPath string) (*OGCStyles, error) {
	var config OGCStyles
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error: %v, could not read config file: %v", err, configPath)
	}
	err = yaml.UnmarshalStrict(content, &config)
	if err != nil {
		return nil, fmt.Errorf("error: %v, could not parse config file: %v", err, configPath)
	}
	config.BaseResource = strings.Trim(config.BaseResource, "/")
	return &config, nil
}

func GenerateDocuments(ogcStyles *OGCStyles, assetDir string, formats []Format) chan *Document {
	documents := make(Documents, documentChanSize)
	go func() {
		defer close(documents)
		styles := Styles{Default: ogcStyles.Default}
		for _, styleMetadata := range ogcStyles.StylesMetadata {
			hasSelf := false
			var stylesLinks []Link
			for i := range styleMetadata.Links {
				styleMetadataLink := &styleMetadata.Links[i]
				err := styleMetadataLink.UpdateHref(ogcStyles.BaseResource, styleMetadata.Id, ogcStyles.AdditionalFormats)
				if err != nil {
					documents.HandleError(fmt.Errorf("error: %s could not update href with base url: %s and id: %s", err, ogcStyles.BaseResource, styleMetadata.Id))
					return
				}
				if styleMetadataLink.Rel == StylesheetRelation {
					log.Printf("warning: stylesheet link found in metadata links %s", *styleMetadataLink.Href)
					continue
				} else if styleMetadataLink.Rel != SelfRelation {
					document, err := generateAssetFromLinkRelation(*styleMetadataLink, styleMetadata.Id, assetDir, ogcStyles)
					if err != nil {
						documents.HandleError(fmt.Errorf("error: %s could not update href with base url: %s and id: %s", err, ogcStyles.BaseResource, styleMetadata.Id))
						return
					} else if document != nil {
						documents <- document
					}
					// OGC API Styles Requirement 3I - If a thumbnail is available for a style in the style metadata (see recommendation /rec/core/style-md-preview), a link with the link relation type preview SHALL also be provided in the Styles resource.
					stylesLinks = append(stylesLinks, *styleMetadataLink)
				} else {
					hasSelf = true
					stylesLinks = append(stylesLinks, styleMetadataLink.WithOtherRelation(DescribedbyRelation))
				}
			}

			if !hasSelf {
				title := fmt.Sprintf("Style Metadata for %s", styleMetadata.Id)
				selfMetadataLink := Link{
					Title: &title,
					Rel:   SelfRelation,
					Href:  DescribedbyRelation.MustToUrl(ogcStyles.BaseResource, styleMetadata.Id),
				}
				styleMetadata.Links = append(styleMetadata.Links, selfMetadataLink)
				// OGC API Styles Requirement 3F Each style SHALL have a link to the style metadata (link relation type: describedby) with the type attribute stating the media type of the metadata encoding.
				stylesLinks = append(stylesLinks, selfMetadataLink.WithOtherRelation(DescribedbyRelation))
			}
			for i := range styleMetadata.Stylesheets {
				stylesheetLink := &styleMetadata.Stylesheets[i].Link
				err := stylesheetLink.UpdateHref(ogcStyles.BaseResource, styleMetadata.Id, ogcStyles.AdditionalFormats)
				if err != nil {
					documents.HandleError(fmt.Errorf("error: %s could not update href with base url: %s and id: %s", err, ogcStyles.BaseResource, styleMetadata.Id))
					return
				}
				document, err := generateAssetFromLinkRelation(*stylesheetLink, styleMetadata.Id, assetDir, ogcStyles)
				// OGC API Styles Requirement 3C - The styles member SHALL include one item for each style currently on the server.
				// OGC API Styles Requirement 3E - Each style SHALL have at least one link to a style encoding supported for the style (link relation type: stylesheet) with the type attribute stating the media type of the style encoding.
				// OGC API Styles Requirement 3H - If a http://www.opengis.net/def/rel/ogc/1.0/schema link to a URI for the schema of the data is available for a style in the style metadata (see recommendation /rec/core/style-md-schema), a link with the same link relation type SHALL also be provided in the Styles resource.
				if err != nil {
					documents.HandleError(fmt.Errorf("error: %s could not update href with base url: %s and id: %s", err, ogcStyles.BaseResource, styleMetadata.Id))
					return
				} else if document != nil {
					documents <- document
				}
				stylesLinks = append(stylesLinks, *stylesheetLink)
			}

			styles.Styles = append(styles.Styles, Style{styleMetadata.Id, *styleMetadata.Title, stylesLinks})
			for _, format := range formats {
				document, err := Render(styleMetadata, DescribedbyRelation.MustToPath(styleMetadata.Id), format)
				if err != nil {
					documents.HandleError(err)
					return
				}
				documents <- document
			}
		}
		for _, format := range formats {
			document, err := Render(styles, stylesResource, format)
			if err != nil {
				documents.HandleError(err)
				return
			}
			documents <- document
		}
	}()
	return documents
}

func generateAssetFromLinkRelation(link Link, styleId string, assetDir string, ogcStyles *OGCStyles) (*Document, error) {
	switch link.Rel {
	case StylesheetRelation, PreviewRelation:
		if link.AssetFilename == nil {
			return nil, fmt.Errorf("asset-filename not specified for stylesheet %s", *link.Href)
		}
		filename := *link.AssetFilename
		assetPath := fmt.Sprintf("%s/%s", assetDir, filename)
		assetContent, err := ioutil.ReadFile(assetPath)
		if err != nil {
			return nil, fmt.Errorf("could not find asset %s", assetPath)
		}

		var contentBuffer bytes.Buffer
		assetTemplate := template.Must(template.New("assetTemplate").Parse(string(assetContent)))
		err = assetTemplate.Execute(&contentBuffer, ogcStyles)
		if err != nil {
			return nil, fmt.Errorf("could not find format asset: %s", assetPath)
		}

		identifier := styleId
		if link.Rel == PreviewRelation {
			identifier = *link.AssetFilename
		}
		path, err := link.Rel.ToPath(identifier)
		return &Document{path, *link.Type, &contentBuffer, nil}, nil
	default:
		log.Printf("not generating asset for link with relation %s, with href %s", link.Rel, *link.Href)
		return nil, nil
	}
}
