# Goas

> A Golang OGC Api Styles Generator

## What it will do

GOAStyles generates [OGC API -
Styles](https://github.com/opengeospatial/ogcapi-styles) ([also see
here](http://docs.opengeospatial.org/DRAFTS/20-009.html)) static documents in
golang. GOAS generates the static files with paths conforming to this spec, and
thus implements the read only aspects of the OGC Styles API. The documents are
generated from a `yaml` config which mirrors the style metadata [like
so](#how-to-configure)

Conformance:

- OGC API styles core
- styles (as is)

Implemented:

- Json format

### Out of Scope

- Serving files
- Conformance to manage-styles and style-validation, since those are dynamic
  endpoints.
- Conformance and core are expected to be implemented elsewhere.

### Wishlist / TODO

- [ ] Implement OGC API spec for layers
- [ ] Azure BlobStore
- [ ] Move beyond json rendering

## How to run

### Locally

```sh
mkdir -p output
go build . && ./goas --file-destination=./output ./examples/assets ./examples/config.yaml
```

#### Building the binary

```sh
go build .
```

#### Testing the binary

```sh
go test ./...
```

#### Usage

Goas supports config options as flags and environment variables (see the
[$VALUES] below) and expects a [config yaml](#how-to-configure)

```bash
NAME:
   Go OGC Api Styles Generator - Generates OGC API styles to S3 or disk

USAGE:
   goas [global options] command [command options] [arguments]

ARGUMENTS:
  [ASSET_DIR]: path that points to directory where the assets (styles, thumbnails) are provided
  [CONFIG]: path to the configuration.yaml for the style generation

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --s3-access-key value     S3 access key (optional) [$S3_ACCESS_KEY]
   --s3-secret value         S3 secret key (optional) [$S3_SECRET_KEY]
   --s3-endpoint value       s3 endpoint with protocol (optional) [$S3_ENDPOINT]
   --s3-bucket value         S3 bucket where the styles land on S3 (optional) [$S3_BUCKET]
   --s3-prefix value         S3 prefix where the styles land on S3 (optional) [$S3_PREFIX]
   --file-destination value  Path where the styles land on disk (optional) [$FILE_DESTINATION]
   --formats value           (stub) comma seperated list of rendered formats. Choose from: [json,] (default: json) [$API_FORMATS]
   --help, -h                show help (default: false)
```

#### How to configure

The main config expects:

```bash
base-resource:      the url that is prepended to each enpdoint (required)
default:            the default style (optional)
additional-formats: key value pairs of custom formats (optional)
styles:             a yaml that conforms to (required); see examples/config.yaml 
                    and examples/minimal_config.yaml for further explanation.
```

## Docker

### docker build

```sh
docker build -t pdok/goas .
```

### docker run local example

```sh
mkdir -p output
docker run --rm -v `pwd`/examples:/examples -v `pwd`/output:/output pdok/goas --file-destination /output /examples/assets /examples/config.yaml
```
