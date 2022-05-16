package pkg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pdok/goas/pkg/models"
	"gopkg.in/yaml.v2"
)

func ParseConfig(configPath string) (*models.OGCStyles, error) {
	var config models.OGCStyles
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

func GenerateDocuments(ogcStyles *models.OGCStyles, assetDir string, formats []models.Format) ([]models.Document, error) {
	var documents []models.Document
	for _, additionalAsset := range ogcStyles.AdditionalAssets {
		assetPathGlob := filepath.Join(assetDir, additionalAsset.Path)
		assetPaths, err := filepath.Glob(assetPathGlob)
		if err != nil {
			return nil, fmt.Errorf("cannot glob %s with error: %s", assetPathGlob, err)
		}
		for _, assetPath := range assetPaths {
			relPath, err := filepath.Rel(assetDir, assetPath)
			if err != nil {
				return nil, fmt.Errorf("cannot take the relative path of %s with error: %s", relPath, err)
			}
			link := models.Link{Rel: models.PreloadRelation, Type: &additionalAsset.MediaType, AssetFilename: &relPath}
			document, err := generateAssetFromLinkRelation(link, "", assetDir, ogcStyles)
			if err != nil {
				return nil, err
			}
			documents = append(documents, *document)
		}
	}
	styles := models.Styles{Default: ogcStyles.Default}
	for _, styleMetadata := range ogcStyles.StylesMetadata {
		var stylesLinks []models.Link
		var selfMetadataLink *models.Link
		for i := range styleMetadata.Links {
			document, link, isSelf, err := generateStyleMetadata(&styleMetadata.Links[i], styleMetadata.Id, assetDir, ogcStyles)
			if err != nil {
				return nil, err
			}
			documents = append(documents, *document)
			stylesLinks = append(stylesLinks, *link)
			if isSelf {
				selfMetadataLink = link
			}
		}

		if selfMetadataLink == nil {
			selfMetadataLink = generateMetadataLink(styleMetadata.Id, ogcStyles)
			styleMetadata.Links = append(styleMetadata.Links, *selfMetadataLink)
			// OGC API Styles Requirement 3F Each style SHALL have a link to the style metadata (link relation type: describedby) with the type attribute stating the media type of the metadata encoding.
			stylesLinks = append(stylesLinks, *selfMetadataLink.WithOtherRelation(models.DescribedbyRelation))
		}
		for i := range styleMetadata.Stylesheets {
			document, err := generateStylesheet(&styleMetadata.Stylesheets[i].Link, styleMetadata.Id, assetDir, ogcStyles)
			if err != nil {
				return nil, err
			}
			documents = append(documents, *document)
			// OGC API Styles Requirement 3C - The styles member SHALL include one item for each style currently on the server.
			stylesLinks = append(stylesLinks, styleMetadata.Stylesheets[i].Link)
		}

		styles.Styles = append(styles.Styles, models.Style{
			Id: styleMetadata.Id, Title: *styleMetadata.Title, Links: stylesLinks,
		})
		for _, format := range formats {
			document, err := Render(styleMetadata, models.DescribedbyRelation.MustToPath(styleMetadata.Id), format)
			if err != nil {
				return nil, err
			}
			documents = append(documents, *document)
		}
	}
	for _, format := range formats {
		document, err := Render(styles, models.StylesResource, format)
		if err != nil {
			return nil, err
		}
		documents = append(documents, *document)
	}
	return documents, nil
}

func generateStyleMetadata(styleMetadataLink *models.Link, metadataId string, assetDir string, styles *models.OGCStyles) (document *models.Document, link *models.Link, hasSelf bool, err error) {
	err = styleMetadataLink.UpdateHref(styles.BaseResource, metadataId, styles.AdditionalFormats, false, true)
	if err != nil {
		return nil, nil, false, fmt.Errorf("error: %s could not update href with base url: %s and id: %s", err, styles.BaseResource, metadataId)
	}
	if styleMetadataLink.Rel == models.StylesheetRelation {
		log.Printf("warning: stylesheet link found in metadata links %s", *styleMetadataLink.Href)
		return nil, nil, false, nil
	} else if styleMetadataLink.Rel != models.SelfRelation {
		document, err = generateAssetFromLinkRelation(*styleMetadataLink, metadataId, assetDir, styles)
		if err != nil {
			return nil, nil, false, fmt.Errorf("error: %s could not update href with base url: %s and id: %s", err, styles.BaseResource, metadataId)
		}
		// OGC API Styles Requirement 3I - If a thumbnail is available for a style in the style metadata (see recommendation /rec/core/style-md-preview), a link with the link relation type preview SHALL also be provided in the Styles resource.
		link = styleMetadataLink
	} else {
		hasSelf = true
		link = styleMetadataLink.WithOtherRelation(models.DescribedbyRelation)
	}
	return document, link, hasSelf, nil
}

func generateStylesheet(stylesheetLink *models.Link, metadataId string, assetDir string, styles *models.OGCStyles) (document *models.Document, err error) {
	err = stylesheetLink.UpdateHref(styles.BaseResource, metadataId, styles.AdditionalFormats, true, false)
	if err != nil {
		return nil, fmt.Errorf("error: %s could not update href with base url: %s and id: %s", err, styles.BaseResource, metadataId)
	}
	document, err = generateAssetFromLinkRelation(*stylesheetLink, metadataId, assetDir, styles)
	// OGC API Styles Requirement 3E - Each style SHALL have at least one link to a style encoding supported for the style (link relation type: stylesheet) with the type attribute stating the media type of the style encoding.
	// OGC API Styles Requirement 3H - If a http://www.opengis.net/def/rel/ogc/1.0/schema link to a URI for the schema of the data is available for a style in the style metadata (see recommendation /rec/core/style-md-schema), a link with the same link relation type SHALL also be provided in the Styles resource.
	if err != nil {
		return nil, fmt.Errorf("error: %s could not update href with base url: %s and id: %s", err, styles.BaseResource, metadataId)
	}
	return document, nil
}

func generateMetadataLink(metadataId string, styles *models.OGCStyles) *models.Link {
	title := fmt.Sprintf("Style Metadata for %s", metadataId)
	selfMetadataLink := models.Link{
		Title: &title,
		Rel:   models.SelfRelation,
		Href:  models.DescribedbyRelation.MustToUrl(styles.BaseResource, metadataId),
	}
	return &selfMetadataLink
}

func generateAssetFromLinkRelation(link models.Link, styleId string, assetDir string, ogcStyles *models.OGCStyles) (*models.Document, error) {
	switch link.Rel {
	case models.StylesheetRelation:
		return generateAssetFromSource(link, styleId, assetDir, ogcStyles, true)
	case models.PreviewRelation, models.PreloadRelation:
		return generateAssetFromSource(link, *link.AssetFilename, assetDir, ogcStyles, false)
	default:
		log.Printf("not generating asset for link with relation %s, with href %s", link.Rel, *link.Href)
		return nil, nil
	}
}

func generateAssetFromSource(link models.Link, identifier string, assetDir string, ogcStyles *models.OGCStyles, useTemplate bool) (*models.Document, error) {
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
	if useTemplate {
		assetTemplate := template.Must(template.New("assetTemplate").Parse(string(assetContent)))
		err = assetTemplate.Execute(&contentBuffer, ogcStyles)
		if err != nil {
			return nil, fmt.Errorf("could not find format asset: %s", assetPath)
		}
	} else {
		contentBuffer = *bytes.NewBuffer(assetContent)
	}

	path, err := link.ToPath(identifier, ogcStyles.AdditionalFormats)
	if err != nil {
		return nil, err
	}
	return &models.Document{Path: path, MediaType: *link.Type, Content: &contentBuffer}, nil
}
