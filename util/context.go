package util

import (
	"fmt"
	"github.com/pdok/goas/pkg"
	"github.com/urfave/cli/v2"
	"strings"
)

type Context struct {
	Writer     Writer
	AssetDir   string
	ConfigPath string
	Formats    []pkg.Format
}

var DefaultFormats = []pkg.Format{pkg.JsonFormat}

func CreateContext(c *cli.Context) (*Context, error) {
	s3Endpoint := c.String("s3-endpoint")
	s3AccessKey := c.String("s3-access-key")
	s3SecretKey := c.String("s3-secret-key")
	s3Bucket := c.String("s3-bucket")
	s3Prefix := c.String("s3-prefix")
	fileDestination := c.String("file-destination")

	var assetDir, configPath string
	if c.NArg() > 1 {
		assetDir = strings.Trim(c.Args().Get(0), "/")
		configPath = strings.Trim(c.Args().Get(1), "/")
	} else {
		return nil, fmt.Errorf("expect ASSET_DIR and CONFIG_PATH as arguments")
	}

	var formats []pkg.Format
	for _, format := range strings.Split(c.String("formats"), ",") {
		if format != "" {
			formats = append(formats, pkg.Format(format))
		}
	}
	if formats == nil {
		formats = DefaultFormats
	}

	writer, err := NewWriter(s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket, s3Prefix, fileDestination)
	if err != nil {
		return nil, err
	}

	return &Context{writer, assetDir, configPath, formats}, nil
}
