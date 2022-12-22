package main

import (
	"github.com/pdok/goas/pkg/models"
	"github.com/urfave/cli/v2"
	"log"
	"os"

	"github.com/pdok/goas/pkg"
	"github.com/pdok/goas/util"
)

func main() {
	app := cli.NewApp()
	app.Name = "Go OGC Api Styles Generator"
	app.Usage = "Generates OGC API styles to local disk or remote object storage (S3 or Azure Blob)"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "s3-access-key",
			Usage:   "S3 access key (optional)",
			EnvVars: []string{"S3_ACCESS_KEY"},
		},
		&cli.StringFlag{
			Name:    "s3-secret",
			Usage:   "S3 secret key (optional)",
			EnvVars: []string{"S3_SECRET_KEY"},
		},
		&cli.StringFlag{
			Name:    "s3-endpoint",
			Usage:   "s3 endpoint with protocol (optional)",
			EnvVars: []string{"S3_ENDPOINT"},
		},
		&cli.StringFlag{
			Name:    "s3-bucket",
			Usage:   "S3 bucket where the styles land on S3 (optional)",
			EnvVars: []string{"S3_BUCKET"},
		},
		&cli.StringFlag{
			Name:    "s3-prefix",
			Usage:   "S3 prefix where the styles land on S3 (optional)",
			EnvVars: []string{"S3_PREFIX"},
		},
		&cli.BoolFlag{
			Name:    "s3-secure",
			Usage:   "use a secure S3 connection [true, false], defaults to false (optional)",
			EnvVars: []string{"S3_SECURE"},
		},
		&cli.StringFlag{
			Name:    "azure-storage-connection-string",
			Usage:   "connection string to Azure Blob storage (optional)",
			EnvVars: []string{"AZURE_STORAGE_CONNECTION_STRING"},
		},
		&cli.StringFlag{
			Name:    "azure-storage-container",
			Usage:   "name of Azure Blob storage container (optional)",
			EnvVars: []string{"AZURE_STORAGE_CONTAINER"},
		},
		&cli.StringFlag{
			Name:    "azure-storage-blobs-prefix",
			Usage:   "Azure Blob key prefix (optional)",
			EnvVars: []string{"BLOBS_PREFIX"},
		},
		&cli.StringFlag{
			Name:    "file-destination",
			Usage:   "Path where the styles land on disk (optional)",
			EnvVars: []string{"FILE_DESTINATION"},
		},
		&cli.StringFlag{
			Name:        "formats",
			Usage:       "(stub) comma seperated list of rendered formats. Choose from: [json,]",
			EnvVars:     []string{"API_FORMATS"},
			DefaultText: models.JsonFormat.Name,
		},
	}
	app.ArgsUsage = "[arguments]\n\nARGUMENTS:\n  [ASSET_DIR]: path that points to directory where the assets (styles, thumbnails) are provided\n  [CONFIG]: path to the configuration.yaml for the style generation"

	app.Action = func(c *cli.Context) error {
		log.Printf("Starting %s...\n", app.Name)

		context, err := util.CreateContext(c)
		if err != nil {
			return err
		}

		err = generate(context)
		if err != nil {
			return err
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func generate(ctx *util.Context) error {
	config, err := pkg.ParseConfig(ctx.ConfigPath)
	if err != nil {
		return err
	}

	err = pkg.Validate(config)
	if err != nil {
		return err
	}

	documents, err := pkg.GenerateDocuments(config, ctx.AssetDir, ctx.Formats)
	if err != nil {
		return err
	}
	writer, err := util.NewWriter(ctx)
	if err != nil {
		return err
	}
	for _, document := range documents {
		err = writer.Write(document.Path, document.Content, document.MediaType)
		if err != nil {
			return err
		}
	}

	return nil
}
