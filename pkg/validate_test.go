package pkg

import (
	"github.com/pdok/goas/pkg/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func ValidOGCStyles() *models.OGCStyles {
	config, _ := ParseConfig("../examples/config.yaml")
	return config
}

func TestValidateValidStyles(t *testing.T) {
	ogcStyles := ValidOGCStyles()
	err := ValidateOGCStyles(ogcStyles)
	require.Nil(t, err)
}

func TestValidateDuplicateStyles(t *testing.T) {
	ogcStyles := ValidOGCStyles()
	expected := "validation errors found: requirement 3D fails; found styles with duplicate ids: night"
	ogcStyles.StylesMetadata = append(ogcStyles.StylesMetadata, ogcStyles.StylesMetadata[0])
	err := ValidateOGCStyles(ogcStyles)
	require.NotNil(t, err)
	require.Equal(t, expected, err.Error())
}

func TestValidateWrongStyleRelation(t *testing.T) {
	ogcStyles := ValidOGCStyles()
	expected := "validation errors found: requirement 3E fails; style night stylesheet definition incorrect"
	ogcStyles.StylesMetadata[0].Stylesheets = []models.StyleSheet{{Link: models.Link{Rel: models.SelfRelation}}}
	err := ValidateOGCStyles(ogcStyles)
	require.NotNil(t, err)
	require.Equal(t, expected, err.Error())
}

func TestValidateUnknownDefaultStyle(t *testing.T) {
	ogcStyles := ValidOGCStyles()
	ogcStyles.Default = "unknown"
	expected := "validation errors found: requirement 3G fails; default  unknown not found in styles"
	err := ValidateOGCStyles(ogcStyles)
	require.NotNil(t, err)
	require.Equal(t, expected, err.Error())
}
