package util

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pdok/goas/pkg/models"
	"log"
	"os"
	"path/filepath"
)

type Writer interface {
	Write(filename string, buffer *bytes.Buffer, mediaType models.MediaType) error
}

type S3Writer struct {
	minioClient *minio.Client
	s3Bucket    string
	s3Prefix    string
	ctx         context.Context
}

type AzureBlobWriter struct {
	blobClient    *azblob.Client
	blobContainer string
	blobPrefix    string
	ctx           context.Context
}

type FileWriter struct {
	FileDestination string
}

func (m S3Writer) Write(filename string, buffer *bytes.Buffer, mediaType models.MediaType) error {
	key := m.s3Prefix + filename
	var opts minio.PutObjectOptions
	if mediaType != "" {
		opts = minio.PutObjectOptions{ContentType: string(mediaType)}
	} else {
		opts = minio.PutObjectOptions{}
	}
	log.Printf("writing to S3: %s with mediaType: %s", key, mediaType)
	_, err := m.minioClient.PutObject(m.ctx, m.s3Bucket, key, buffer, int64(buffer.Len()), opts)
	if err != nil {
		return fmt.Errorf("error: %s, could not write file %s to S3", err, filename)
	}
	return nil
}

func (m AzureBlobWriter) Write(filename string, buffer *bytes.Buffer, mediaType models.MediaType) error {
	key := m.blobPrefix + filename
	var opts azblob.UploadBufferOptions
	if mediaType != "" {
		contentType := string(mediaType)
		opts = azblob.UploadBufferOptions{HTTPHeaders: &blob.HTTPHeaders{BlobContentType: &contentType}}
	} else {
		opts = azblob.UploadBufferOptions{}
	}
	log.Printf("writing to Azure Blob: %s with mediaType: %s", key, mediaType)
	_, err := m.blobClient.UploadBuffer(m.ctx, m.blobContainer, key, buffer.Bytes(), &opts)
	if err != nil {
		return fmt.Errorf("error: %s, could not write file %s to Azure Blob", err, filename)
	}
	return nil
}

func (f FileWriter) makeDirIfNotExists(path string) error {
	dir, _ := filepath.Split(path)
	err := os.MkdirAll(dir, os.ModePerm)

	if err == nil || os.IsExist(err) {
		return nil
	} else {
		return err
	}
}

func (f FileWriter) Write(path string, buffer *bytes.Buffer, _ models.MediaType) error {
	filename := filepath.Join(f.FileDestination, path)
	err := f.makeDirIfNotExists(filename)
	if err != nil {
		return fmt.Errorf("could not make dir for: %s : %s ", filename, err.Error())
	}
	if _, err := os.Stat(filename); os.IsExist(err) {
		if err != nil {
			return err
		}
	}
	log.Printf("writing: %s", filename)
	fileWriter, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could create file %s : %s ", filename, err.Error())
	}
	_, err = buffer.WriteTo(fileWriter)
	if err != nil {
		return fmt.Errorf("could not write to file %s : %s ", filename, err.Error())
	}
	return nil
}

func newS3Writer(s3Endpoint string, s3AccessKey string, s3SecretKey string, s3Bucket string, s3Prefix string, s3Secure bool) (Writer, error) {
	minioClient, err := minio.New(s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3AccessKey, s3SecretKey, ""),
		Secure: s3Secure,
	})
	if err != nil {
		return nil, err
	}
	return &S3Writer{
		minioClient,
		s3Bucket,
		s3Prefix,
		context.Background(),
	}, nil
}

func newAzureBlobWriter(connectionString string, container string, prefix string) (Writer, error) {
	blobClient, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, err
	}
	return &AzureBlobWriter{
		blobClient,
		container,
		prefix,
		context.Background(),
	}, nil
}

func NewWriter(ctx *Context) (writer Writer, err error) {
	if ctx.StorageDestination == FILE {
		writer = &FileWriter{*ctx.FileDestination}
	} else if ctx.StorageDestination == S3 {
		writer, err = newS3Writer(ctx.S3.Endpoint, ctx.S3.AccessKey, ctx.S3.SecretKey, ctx.S3.Bucket, ctx.S3.Prefix, ctx.S3.Secure)
		if err != nil {
			return nil, err
		}
	} else if ctx.StorageDestination == AZURE_BLOB {
		writer, err = newAzureBlobWriter(ctx.AzureBlob.ConnectionString, ctx.AzureBlob.Container, ctx.AzureBlob.Prefix)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid storage destination provided")
	}
	return writer, nil
}
