FROM golang:1.19-alpine3.16 AS build

# Create and change to the build directory.
WORKDIR /build

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download


COPY . /build

#disable crosscompiling
ENV CGO_ENABLED=0

#compile linux only
ENV GOOS=linux

RUN CGO_ENABLED=0 go build
RUN go test -short ./...

FROM scratch

WORKDIR /
ENV PATH=/

COPY --from=build /build/goas /
ENTRYPOINT ["/goas"]
