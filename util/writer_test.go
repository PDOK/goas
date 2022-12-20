package util

import (
	"bytes"
	gocontext "context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/docker/go-connections/nat"
	"github.com/pdok/goas/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
)

const bucket = "public-test"

func TestWriteToAzureBlob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// given
	ctx := gocontext.Background()
	expected := []byte("This is an example blob on Azure")

	port, container, err := setupAzurite(t, ctx)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate container: %s", err.Error())
		}
	}()
	blobClient := setupBlobs(t, port, ctx)

	storageDest, _, _, azureBlobContext, err := initStorage("", "", "", "", "", "", false, getConnectionString(port), bucket, "foo")
	if err != nil {
		t.Fatalf("Failed to init storage")
	}
	writer, err := NewWriter(&Context{nil, &azureBlobContext, nil, storageDest, "", "", nil})
	if err != nil {
		t.Fatalf("Failed to init writer")
	}

	// when
	writer.Write("bar.json", bytes.NewBuffer(expected), models.JsonMediaType)

	// then
	var actual = make([]byte, len(expected))
	_, err = blobClient.DownloadBuffer(ctx, bucket, "foo/bar.json", actual, nil)
	if err != nil {
		t.Fatalf("error %v", err)
	}
	assert.Equal(t, expected, actual)
}

func setupBlobs(t *testing.T, port nat.Port, ctx gocontext.Context) *azblob.Client {
	blobClient, err := azblob.NewClientFromConnectionString(getConnectionString(port), nil)
	if err != nil {
		t.Error(err)
	}
	_, err = blobClient.CreateContainer(ctx, bucket, nil)
	if err != nil {
		t.Error(err)
	}
	return blobClient
}

func getConnectionString(port nat.Port) string {
	return fmt.Sprintf("BlobEndpoint=http://127.0.0.1:%d/devstoreaccount1;"+
		"DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;"+
		"AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;",
		port.Int())
}

func setupAzurite(t *testing.T, ctx gocontext.Context) (nat.Port, testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mcr.microsoft.com/azure-storage/azurite:latest",
		ExposedPorts: []string{"10000/tcp"},
		Cmd:          []string{"azurite-blob", "--blobHost", "0.0.0.0"},
		WaitingFor:   wait.ForLog("Azurite Blob service successfully listens"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	port, err := container.MappedPort(ctx, "10000/tcp")
	if err != nil {
		t.Error(err)
	}
	return port, container, err
}
