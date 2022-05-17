package util

import (
	"errors"
	"fmt"
	"github.com/pdok/goas/pkg/models"
	"github.com/urfave/cli/v2"
	"strings"
)

type S3Context struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Prefix    string
	Secure    bool
}

type Context struct {
	S3              *S3Context
	FileDestination *string
	isLocal         bool
	AssetDir        string
	ConfigPath      string
	Formats         []models.Format
}

var DefaultFormats = []models.Format{models.JsonFormat}

func CreateContext(c *cli.Context) (*Context, error) {
	s3Endpoint := c.String("s3-endpoint")
	s3AccessKey := c.String("s3-access-key")
	s3SecretKey := c.String("s3-secret")
	s3Bucket := c.String("s3-bucket")
	s3Prefix := c.String("s3-prefix")
	s3Secure := c.Bool("s3-secure")
	fileDestination := c.String("file-destination")

	isLocal := fileDestination != ""
	isS3 := s3Endpoint != "" && s3SecretKey != "" && s3Bucket != "" && s3AccessKey != "" && s3Prefix != ""
	if (isS3 && isLocal) || (!isS3 && !isLocal) {
		return nil, errors.New("provide either valid S3 configuration, or a local file destination")
	}

	var s3Context S3Context
	var fileDest *string
	if isLocal {
		fileDest = &fileDestination
	} else {
		if !strings.HasSuffix(s3Prefix, "/") {
			s3Prefix = s3Prefix + "/"
		}
		s3Context = S3Context{s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket, s3Prefix, s3Secure}
	}

	var assetDir, configPath string
	if c.NArg() > 1 {
		assetDir = strings.Trim(c.Args().Get(0), "/")
		configPath = strings.Trim(c.Args().Get(1), "/")
	} else {
		return nil, fmt.Errorf("expect ASSET_DIR and CONFIG_PATH as arguments")
	}

	var formats []models.Format
	for _, format := range strings.Split(c.String("formats"), ",") {
		if format != "" {
			f, ok := models.GetFormat(format)
			if ok {
				formats = append(formats, f)
			}
		}
	}
	if formats == nil {
		formats = DefaultFormats
	}

	return &Context{&s3Context, fileDest, isLocal, assetDir, configPath, formats}, nil
}
