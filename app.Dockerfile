# syntax=docker/dockerfile:1
# adapted from https://docs.docker.com/language/golang/build-images/#multi-stage-builds

# Build the application from source
FROM golang:1.21-alpine AS build-stage

WORKDIR /app

COPY . .
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o binary

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /app/binary /binary

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/binary"]
