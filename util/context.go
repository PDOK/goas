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

type AzureBlobContext struct {
	ConnectionString string
	Container        string
	Prefix           string
}

type Context struct {
	S3                 *S3Context
	AzureBlob          *AzureBlobContext
	FileDestination    *string
	StorageDestination StorageDestination
	AssetDir           string
	ConfigPath         string
	Formats            []models.Format
}

type StorageDestination string

const (
	FILE       StorageDestination = "FILE"
	S3         StorageDestination = "S3"
	AZURE_BLOB StorageDestination = "AZURE_BLOB"
)

var DefaultFormats = []models.Format{models.JsonFormat}

func CreateContext(c *cli.Context) (*Context, error) {
	storageDest, fileDest, s3Context, azureBlobContext, err := initStorage(
		c.String("file-destination"),
		c.String("s3-endpoint"),
		c.String("s3-secret"),
		c.String("s3-bucket"),
		c.String("s3-access-key"),
		c.String("s3-prefix"),
		c.Bool("s3-secure"),
		c.String("azure-storage-connection-string"),
		c.String("azure-storage-container"),
		c.String("azure-storage-blobs-prefix"))
	if err != nil {
		return nil, err
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

	return &Context{&s3Context, &azureBlobContext, fileDest,
		storageDest, assetDir, configPath, formats}, nil
}

func initStorage(fileDestination string, s3Endpoint string, s3SecretKey string, s3Bucket string,
	s3AccessKey string, s3Prefix string, s3Secure bool, azureConnectionString string,
	azureContainer string, azurePrefix string) (StorageDestination, *string, S3Context, AzureBlobContext, error) {

	var s3Context S3Context
	var azureBlobContext AzureBlobContext
	var fileDest *string

	var storageDestination StorageDestination
	if fileDestination != "" {
		storageDestination = FILE
		fileDest = &fileDestination
	} else if s3Endpoint != "" && s3SecretKey != "" && s3Bucket != "" && s3AccessKey != "" && s3Prefix != "" {
		storageDestination = S3
		if !strings.HasSuffix(s3Prefix, "/") {
			s3Prefix = s3Prefix + "/"
		}
		s3Context = S3Context{s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket, s3Prefix, s3Secure}
	} else if azureConnectionString != "" && azureContainer != "" && azurePrefix != "" {
		storageDestination = AZURE_BLOB
		if !strings.HasSuffix(azurePrefix, "/") {
			azurePrefix = azurePrefix + "/"
		}
		azureBlobContext = AzureBlobContext{azureConnectionString, azureContainer, azurePrefix}
	} else {
		return "", nil, S3Context{}, AzureBlobContext{}, errors.New("provide either a valid file destination, S3 config or Azure Blob config")
	}
	return storageDestination, fileDest, s3Context, azureBlobContext, nil
}
