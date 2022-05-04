package pkg

import (
	"github.com/pdok/goas/pkg/models"
	"github.com/stretchr/testify/require"
	"testing"
)

type TestStruct struct {
	NestedTest NestedTestStruct `json:"nested" yaml:"nested"`
}

type NestedTestStruct struct {
	Test string `json:"test" yaml:"test"`
}

func TestRenderJson(t *testing.T) {
	obj := TestStruct{NestedTestStruct{"test"}}
	path := "test.json"
	expected := "{\"nested\":{\"test\":\"test\"}}\n"

	result, err := Render(obj, path, models.JsonFormat)

	require.Nil(t, err)
	require.Equal(t, models.JsonMediaType, result.MediaType)
	require.Equal(t, expected, result.Content.String())
	require.Equal(t, result.Path, path)
}
