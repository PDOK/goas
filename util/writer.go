package util

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pdok/goas/pkg"
	"log"
	"os"
	"path/filepath"
)

type Writer interface {
	Write(filename string, buffer *bytes.Buffer, mediaType pkg.MediaType, additionalFormats map[pkg.MediaType]pkg.Format) error
}

type MinioWriter struct {
	minioClient *minio.Client
	s3Bucket    string
	s3Prefix    string
	ctx         context.Context
}

type FileWriter struct {
	FileDestination string
}

func (m MinioWriter) Write(filename string, buffer *bytes.Buffer, mediaType pkg.MediaType, _ map[pkg.MediaType]pkg.Format) error {
	key := m.s3Prefix + filename
	var opts minio.PutObjectOptions
	if mediaType == "" {
		opts = minio.PutObjectOptions{ContentType: string(mediaType)}
	} else {
		opts = minio.PutObjectOptions{}
	}
	log.Printf("writing to S3: %s with mediaType: %s", key, mediaType)
	_, err := m.minioClient.PutObject(m.ctx, m.s3Bucket, key, buffer, int64(buffer.Len()), opts)
	if err != nil {
		return fmt.Errorf("error: %s, could not write file %s to minio", err, filename)
	}
	return nil
}

func (f FileWriter) makeDirIfNotExists(path string) error {
	dir, _ := filepath.Split(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f FileWriter) Write(path string, buffer *bytes.Buffer, mediaType pkg.MediaType, additionalFormats map[pkg.MediaType]pkg.Format) error {
	extension := mediaType.ToFormat(additionalFormats, false)
	if extension != "" {
		path = fmt.Sprintf("%s.%s", path, extension)
	}
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

func newMinioWriter(s3Endpoint string, s3AccessKey string, s3SecretKey string, s3Bucket string, s3Prefix string) (Writer, error) {
	minioClient, err := minio.New(s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3AccessKey, s3SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	return &MinioWriter{
		minioClient,
		s3Bucket,
		s3Prefix,
		context.Background(),
	}, nil
}

func NewWriter(s3Endpoint string, s3AccessKey string, s3SecretKey string, s3Bucket string, s3Prefix string, fileDestination string) (writer Writer, err error) {
	isLocal := fileDestination != ""
	isS3 := s3Endpoint != "" && s3SecretKey != "" && s3Bucket != "" && s3AccessKey != "" && s3Prefix != ""
	if (isS3 && isLocal) || (!isS3 && !isLocal) {
		return nil, errors.New("provide either valid S3 configuration, or a local file destination")
	}

	if isLocal {
		writer = &FileWriter{fileDestination}
	} else {
		writer, err = newMinioWriter(s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket, s3Prefix)
		if err != nil {
			return nil, err
		}
	}
	return writer, nil
}
