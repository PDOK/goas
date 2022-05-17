package pkg

import (
	"github.com/pdok/goas/pkg/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func ValidStyles() *models.StylesConfig {
	config, _ := ParseConfig("../examples/config.yaml")
	return config
}

func TestValidateValidStyles(t *testing.T) {
	stylesConfig := ValidStyles()
	err := Validate(stylesConfig)
	require.Nil(t, err)
}

func TestValidateDuplicateStyles(t *testing.T) {
	stylesConfig := ValidStyles()
	expected := "validation errors found: requirement 3D fails; found styles with duplicate ids: night"
	stylesConfig.StylesMetadata = append(stylesConfig.StylesMetadata, stylesConfig.StylesMetadata[0])
	err := Validate(stylesConfig)
	require.NotNil(t, err)
	require.Equal(t, expected, err.Error())
}

func TestValidateWrongStyleRelation(t *testing.T) {
	stylesConfig := ValidStyles()
	expected := "validation errors found: requirement 3E fails; style night stylesheet definition incorrect"
	stylesConfig.StylesMetadata[0].Stylesheets = []models.StyleSheet{{Link: models.Link{Rel: models.SelfRelation}}}
	err := Validate(stylesConfig)
	require.NotNil(t, err)
	require.Equal(t, expected, err.Error())
}

func TestValidateUnknownDefaultStyle(t *testing.T) {
	stylesConfig := ValidStyles()
	stylesConfig.Default = "unknown"
	expected := "validation errors found: requirement 3G fails; default  unknown not found in styles"
	err := Validate(stylesConfig)
	require.NotNil(t, err)
	require.Equal(t, expected, err.Error())
}
