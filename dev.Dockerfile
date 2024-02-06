# syntax=docker/dockerfile:1
# adapted from https://docs.docker.com/language/golang/build-images/#multi-stage-builds

# Build the application from source
FROM golang:1.21-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download

COPY *.go ./
